package com.ev.battery.agent;

import java.io.IOException;
import java.nio.file.Files;
import java.nio.file.Path;
import java.util.Map;

import dev.langchain4j.data.document.loader.FileSystemDocumentLoader;
import dev.langchain4j.data.document.splitter.DocumentSplitters;
import dev.langchain4j.data.segment.TextSegment;
import dev.langchain4j.memory.chat.MessageWindowChatMemory;
import dev.langchain4j.model.vertexai.VertexAiEmbeddingModel;
import dev.langchain4j.model.vertexai.VertexAiGeminiChatModel;
import dev.langchain4j.rag.content.retriever.ContentRetriever;
import dev.langchain4j.rag.content.retriever.EmbeddingStoreContentRetriever;
import dev.langchain4j.service.AiServices;
import dev.langchain4j.store.embedding.EmbeddingStoreIngestor;
import dev.langchain4j.store.embedding.inmemory.InMemoryEmbeddingStore;

/**
 * Initializes the expensive shared resources (models, per-model embedding stores) once at startup.
 *
 * Vehicle routing:
 *   R1S         → docs/R1S/  (R1S owner manuals only)
 *   R1T         → docs/R1T/  (R1T owner manuals only)
 *   UNKNOWN/null → falls back to R1S (avoids embedding all docs a third time)
 *
 * Use newAgent(vehicleModel) to get a fresh EvExpert with the correct retriever and clean memory.
 */
public class AgentFactory {
    private final VertexAiGeminiChatModel chatModel;
    private final Map<String, ContentRetriever> retrievers;

    public AgentFactory(String projectId, String location) {
        this.chatModel = VertexAiGeminiChatModel.builder()
            .project(projectId)
            .location(location)
            .modelName("gemini-2.0-flash")
            .temperature(0.0f)
            .topP(0.95f)
            .build();

        // maxSegmentsPerBatch: text-embedding-004 caps at 20k tokens/batch.
        // At 300 tokens/chunk, 50 segments = ~15,000 tokens/batch — safely under the limit.
        // Higher batch size = fewer API round-trips = faster initial embedding.
        VertexAiEmbeddingModel embeddingModel = VertexAiEmbeddingModel.builder()
            .project(projectId)
            .location(location)
            .modelName("text-embedding-004")
            .publisher("google")
            .maxSegmentsPerBatch(50)
            .build();

        ContentRetriever r1sRetriever = buildRetriever(embeddingModel, "docs/R1S");
        ContentRetriever r1tRetriever = buildRetriever(embeddingModel, "docs/R1T");
        this.retrievers = Map.of(
            "R1S", r1sRetriever,
            "R1T", r1tRetriever
        );
    }

    /**
     * Returns a fresh agent wired to the correct vehicle's document store.
     * Unrecognized or null vehicleModel falls back to R1S.
     */
    public EvExpert newAgent(String vehicleModel) {
        String key = vehicleModel != null ? vehicleModel.toUpperCase() : "";
        ContentRetriever retriever = retrievers.getOrDefault(key, retrievers.get("R1S"));
        return AiServices.builder(EvExpert.class)
            .chatLanguageModel(chatModel)
            .contentRetriever(retriever)
            .chatMemory(MessageWindowChatMemory.withMaxMessages(10))
            .tools(new JiraTicketingTool())
            .build();
    }

    /**
     * Builds a retriever for the given docs path, using a cached embedding store if available.
     * Cache is stored in embeddings/<name>.json (e.g. embeddings/R1S.json).
     * First run embeds and saves; subsequent runs load from disk instantly.
     */
    private ContentRetriever buildRetriever(VertexAiEmbeddingModel embeddingModel, String docsPath) {
        String cacheName = docsPath.replace("/", "_").replace("\\", "_");
        Path cacheFile = Path.of("embeddings", cacheName + ".json");

        InMemoryEmbeddingStore<TextSegment> store;
        if (Files.exists(cacheFile)) {
            System.out.println("Loading embedding cache: " + cacheFile);
            try {
                store = InMemoryEmbeddingStore.fromFile(cacheFile);
            } catch (Exception e) {
                System.out.println("Cache load failed, re-embedding: " + e.getMessage());
                store = embedAndSave(embeddingModel, docsPath, cacheFile);
            }
        } else {
            System.out.println("No cache found for " + docsPath + " — embedding now (one-time)...");
            store = embedAndSave(embeddingModel, docsPath, cacheFile);
        }

        return EmbeddingStoreContentRetriever.builder()
            .embeddingModel(embeddingModel)
            .embeddingStore(store)
            .maxResults(5)
            .build();
    }

    private InMemoryEmbeddingStore<TextSegment> embedAndSave(
            VertexAiEmbeddingModel embeddingModel, String docsPath, Path cacheFile) {
        InMemoryEmbeddingStore<TextSegment> store = new InMemoryEmbeddingStore<>();
        EmbeddingStoreIngestor.builder()
            .documentSplitter(DocumentSplitters.recursive(300, 30))
            .embeddingModel(embeddingModel)
            .embeddingStore(store)
            .build()
            .ingest(FileSystemDocumentLoader.loadDocuments(docsPath));
        try {
            Files.createDirectories(cacheFile.getParent());
            store.serializeToFile(cacheFile);
            System.out.println("Embedding cache saved: " + cacheFile);
        } catch (IOException e) {
            System.out.println("Warning: could not save embedding cache: " + e.getMessage());
        }
        return store;
    }
}
