package com.ev.battery.agent;

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
 * Initializes the expensive shared resources (models, embedding store) once at startup.
 * Use newAgent() to get a fresh EvExpert instance with clean chat memory per request.
 */
public class AgentFactory {
    private final VertexAiGeminiChatModel chatModel;
    private final ContentRetriever retriever;

    public AgentFactory(String projectId, String location) {
        this.chatModel = VertexAiGeminiChatModel.builder()
            .project(projectId)
            .location(location)
            .modelName("gemini-2.0-flash")
            .temperature(0.0f)
            .topP(0.95f)
            .build();

        VertexAiEmbeddingModel embeddingModel = VertexAiEmbeddingModel.builder()
            .project(projectId)
            .location(location)
            .modelName("text-embedding-004")
            .publisher("google")
            .build();

        InMemoryEmbeddingStore<TextSegment> store = new InMemoryEmbeddingStore<>();
        EmbeddingStoreIngestor.builder()
            .documentSplitter(DocumentSplitters.recursive(500, 50))
            .embeddingModel(embeddingModel)
            .embeddingStore(store)
            .build()
            .ingest(FileSystemDocumentLoader.loadDocuments("docs/"));

        this.retriever = EmbeddingStoreContentRetriever.builder()
            .embeddingModel(embeddingModel)
            .embeddingStore(store)
            .maxResults(5)
            .build();
    }

    /** Returns a fresh agent with clean chat memory. Safe to call per-request. */
    public EvExpert newAgent() {
        return AiServices.builder(EvExpert.class)
            .chatLanguageModel(chatModel)
            .contentRetriever(retriever)
            .chatMemory(MessageWindowChatMemory.withMaxMessages(10))
            .tools(new JiraTicketingTool())
            .build();
    }
}
