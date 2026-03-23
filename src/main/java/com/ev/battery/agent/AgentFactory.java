package com.ev.battery.agent;

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
 *   R1S → docs/R1S/  (R1S owner manuals only)
 *   R1T → docs/R1T/  (R1T owner manuals only)
 *   UNKNOWN → docs/  (all documents, fallback)
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
        // At 300 tokens/chunk, 20 segments = ~6,000 tokens/batch — safely under the limit.
        VertexAiEmbeddingModel embeddingModel = VertexAiEmbeddingModel.builder()
            .project(projectId)
            .location(location)
            .modelName("text-embedding-004")
            .publisher("google")
            .maxSegmentsPerBatch(20)
            .build();

        this.retrievers = Map.of(
            "R1S", buildRetriever(embeddingModel, "docs/R1S"),
            "R1T", buildRetriever(embeddingModel, "docs/R1T"),
            "UNKNOWN", buildRetriever(embeddingModel, "docs")
        );
    }

    /**
     * Returns a fresh agent wired to the correct vehicle's document store.
     * vehicleModel should be "R1S", "R1T", or "UNKNOWN".
     */
    public EvExpert newAgent(String vehicleModel) {
        ContentRetriever retriever = retrievers.getOrDefault(
            vehicleModel != null ? vehicleModel.toUpperCase() : "UNKNOWN",
            retrievers.get("UNKNOWN")
        );
        return AiServices.builder(EvExpert.class)
            .chatLanguageModel(chatModel)
            .contentRetriever(retriever)
            .chatMemory(MessageWindowChatMemory.withMaxMessages(10))
            .tools(new JiraTicketingTool())
            .build();
    }

    private ContentRetriever buildRetriever(VertexAiEmbeddingModel embeddingModel, String docsPath) {
        InMemoryEmbeddingStore<TextSegment> store = new InMemoryEmbeddingStore<>();
        EmbeddingStoreIngestor.builder()
            .documentSplitter(DocumentSplitters.recursive(300, 30))
            .embeddingModel(embeddingModel)
            .embeddingStore(store)
            .build()
            .ingest(FileSystemDocumentLoader.loadDocuments(docsPath));
        return EmbeddingStoreContentRetriever.builder()
            .embeddingModel(embeddingModel)
            .embeddingStore(store)
            .maxResults(5)
            .build();
    }
}
