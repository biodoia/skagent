# SkAgent - Code Review & Piano di Migliorie

## 📋 Code Review Summary

### ✅ Punti di Forza
1. **Architettura modulare**: Buona separazione tra TUI, AI, tools e config
2. **Multi-provider support**: OpenRouter, Claude Max, Gemini CLI, DeepSeek, etc.
3. **Setup wizard**: Esperienza utente guidata per la configurazione
4. **Makefile completo**: Build, cross-compile, install targets
5. **Uso di Charm stack**: BubbleTea, Bubbles, Lipgloss - scelte moderne

### ⚠️ Problemi Identificati

#### 1. **Mancanza di Context nel Tool Execution**
- I tools non usano `context.Context` - impossibile cancellare operazioni
- Nessun timeout sui comandi CLI

#### 2. **Web Search Tool Non Implementato**
- Solo placeholder, nessuna reale funzionalità

#### 3. **Nessun Streaming delle Risposte**
- Le risposte AI arrivano tutte insieme, UX lenta

#### 4. **Error Handling Incompleto**
- Molti errori vengono solo loggati, non gestiti
- Nessun retry logic per chiamate API

#### 5. **Mancanza di Tests**
- Zero test unitari o integration tests

#### 6. **Documentazione Incompleta**
- Nessun README.md
- Nessun esempio di utilizzo

#### 7. **Tool Calls Non Integrati**
- `CompleteWithTools` non viene usato nel flusso principale

#### 8. **Dipendenze Deprecate**
- `ioutil.ReadDir` deprecato, usare `os.ReadDir`

---

## 🚀 Piano di Migliorie

### Fase 1: Bug Fixes & Stabilità (Priorità Alta)

#### 1.1 Aggiungere Context ai Tools
```go
type Tool interface {
    Name() string
    Description() string
    Execute(ctx context.Context, input string) (string, error)
    CanHandle(intent string) bool
}
```

#### 1.2 Fix Deprecazioni
- Sostituire `ioutil.ReadDir` con `os.ReadDir`
- Aggiornare import paths

#### 1.3 Aggiungere Timeout & Retry
- Timeout configurabile per chiamate API
- Retry con exponential backoff

### Fase 2: Nuove Funzionalità (Priorità Media)

#### 2.1 Implementare Web Search
- Integrazione con DuckDuckGo API (gratuita)
- Ricerca GitHub repos via API
- Cache risultati

#### 2.2 Streaming Responses
- Implementare SSE per OpenRouter
- Aggiornamento progressivo UI

#### 2.3 Tool Calling Attivo
- Connettere `CompleteWithTools` al flusso principale
- Loop automatico per tool execution

#### 2.4 Logging Strutturato
- Aggiungere slog per logging
- File di log opzionale

### Fase 3: Testing & Docs (Priorità Media-Alta)

#### 3.1 Unit Tests
- Tests per ogni package
- Mock per chiamate API

#### 3.2 README.md Completo
- Installazione
- Configurazione
- Esempi
- Screenshots

### Fase 4: Features Avanzate (Priorità Bassa)

#### 4.1 MCP Integration
- Supporto MCP servers
- gh-mcp-server integration nativa

#### 4.2 Persistent History
- Salvataggio conversazioni
- Resume sessioni

#### 4.3 Plugin System
- Caricamento dinamico tools
- API per third-party tools

---

## 📁 File da Modificare/Creare

### Modifiche:
1. `internal/tools/interface.go` - Aggiungere Context
2. `internal/tools/speckit.go` - Usare Context, timeout
3. `internal/tools/github.go` - Usare Context, timeout
4. `internal/tools/websearch.go` - Implementare reale
5. `internal/docs/loader.go` - Fix deprecazioni
6. `internal/ai/client.go` - Streaming support
7. `internal/ai/providers.go` - Streaming, retry logic
8. `internal/tui/app.go` - Tool calling loop, streaming

### Nuovi File:
1. `README.md` - Documentazione completa
2. `internal/tools/websearch_impl.go` - DuckDuckGo integration
3. `internal/ai/streaming.go` - Streaming handler
4. `internal/retry/retry.go` - Retry logic
5. `tests/*_test.go` - Test files

---

## ⏱️ Timeline Stimata

| Fase | Tempo | Descrizione |
|------|-------|-------------|
| 1.1-1.3 | 1h | Bug fixes & stabilità |
| 2.1 | 1h | Web search implementation |
| 2.2-2.3 | 2h | Streaming & tool calling |
| 3.1-3.2 | 1h | README & basic tests |

**Totale stimato: ~5 ore per MVP solido**

---

## 🎯 Prossimi Passi Immediati

1. ✅ Creare README.md
2. ✅ Fix deprecazioni in docs/loader.go
3. ✅ Aggiungere Context ai tools
4. ✅ Implementare web search base (GitHub + DuckDuckGo)
5. ✅ Aggiungere retry logic con exponential backoff
6. ✅ Aggiungere unit tests
7. ✅ Fix quote extraction bug
8. ✅ Push su GitHub

## 📊 Stato Attuale

- **Repository**: https://github.com/biodoia/skagent
- **Tests**: ✅ Tutti passano
- **Build**: ✅ Compila senza errori
- **Documentazione**: ✅ README completo
