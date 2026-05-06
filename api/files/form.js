const form = document.getElementById("ia-form");
const engineEl = document.getElementById("engine");
const modelEl = document.getElementById("model");
const promptEl = document.getElementById("prompt");
const pdfInput = document.getElementById("pdf");
const pdfViewer = document.getElementById("pdf-viewer");
const togglePdfBtn = document.getElementById("toggle-pdf");
const reqLoader = document.getElementById("req-loader");

const loaderWrap = document.getElementById("loader-wrap");
const idleText = document.getElementById("idle-text");
const statusText = document.getElementById("status-text");

const inputTokenCountEl = document.getElementById("input-token-count");
const outputTokenCountEl = document.getElementById("output-token-count");
const resultEl = document.getElementById("result");
const submitBtn = document.getElementById("submit-btn");

const chatForm = document.getElementById("chat-form");
const chatInputEl = document.getElementById("chat-input");
const chatMessagesEl = document.getElementById("chat-messages");
const chatSubmitBtn = document.getElementById("chat-submit-btn");
const chatStatusEl = document.getElementById("chat-status");

let pdfObjectUrl = null;
let currentDocumentId = null;
const CHAT_STORAGE_PREFIX = "chat_history:";

const MODELS_BY_ENGINE = {
    AWS: [
        { value: "anthropic.claude-sonnet-4-6", label: "Claude Sonnet 4.6" },
        { value: "anthropic.claude-haiku-4-5-20251001-v1:0", label: "Claude Haiku 4.5" },
        { value: "amazon.nova-pro-v1:0", label: "Amazon Nova Pro" },
        { value: "cohere.command-r-plus-v1:0", label: "Cohere Command R+" }
    ],
    OLLAMA: [
        { value: "gemma2:9b", label: "Gemma 2 9B" },
        { value: "llama3.1:8b", label: "Llama 3.1 8B" },
        { value: "gemma4:e2b", label: "Gemma 4 E2B" },
    ]
};

function estimateTokens(text) {
    if (!text || !text.trim()) return 0;
    return Math.ceil(text.trim().length / 4);
}

function updateInputTokenCount() {
    if (!promptEl || !inputTokenCountEl) return;
    const count = estimateTokens(promptEl.value);
    inputTokenCountEl.textContent = `Tokens input (aprox): ${count}`;
}

function setLoading(isLoading, message = "Enviando solicitud...") {
    if (!loaderWrap || !idleText || !statusText || !reqLoader) return;
    loaderWrap.hidden = !isLoading;
    idleText.hidden = isLoading;
    statusText.textContent = message;
    reqLoader.hidden = !isLoading;
}

function populateModels(engine) {
    modelEl.innerHTML = "";

    if (!engine || !MODELS_BY_ENGINE[engine]) {
        const option = document.createElement("option");
        option.value = "";
        option.textContent = "Primero selecciona un engine";
        modelEl.appendChild(option);
        modelEl.disabled = true;
        return;
    }

    const placeholder = document.createElement("option");
    placeholder.value = "";
    placeholder.textContent = "Selecciona un modelo";
    modelEl.appendChild(placeholder);

    MODELS_BY_ENGINE[engine].forEach((m) => {
        const option = document.createElement("option");
        option.value = m.value;
        option.textContent = m.label;
        modelEl.appendChild(option);
    });

    modelEl.disabled = false;
    modelEl.value = "";
}

function appendChatMessage(role, text) {
    if (!chatMessagesEl) return;

    const empty = chatMessagesEl.querySelector(".chat-empty");
    if (empty) empty.remove();

    const msg = document.createElement("div");
    msg.className = `chat-msg ${role}`;
    msg.innerHTML = `<strong>${role === "user" ? "Tú" : "Asistente"}</strong>${marked.parse(text || "")}`;
    chatMessagesEl.appendChild(msg);
    chatMessagesEl.scrollTop = chatMessagesEl.scrollHeight;
}

function chatStorageKey(documentId) {
    return `${CHAT_STORAGE_PREFIX}${documentId}`;
}

function loadChatHistory(documentId) {
    if (!documentId) return [];

    try {
        const raw = localStorage.getItem(chatStorageKey(documentId));
        if (!raw) return [];
        const parsed = JSON.parse(raw);
        if (!Array.isArray(parsed)) return [];

        return parsed
            .map((m) => ({
                role: String(m.role || "").toLowerCase(),
                content: String(m.content || "")
            }))
            .filter((m) => (m.role === "user" || m.role === "assistant") && m.content.trim() !== "");
    } catch {
        return [];
    }
}

function saveChatHistory(documentId, messages) {
    if (!documentId) return;
    localStorage.setItem(chatStorageKey(documentId), JSON.stringify(messages));
}

function renderChatHistory(messages) {
    if (!chatMessagesEl) return;

    chatMessagesEl.innerHTML = "";
    if (!messages.length) {
        chatMessagesEl.innerHTML = '<p class="chat-empty">Primero genera un resumen y luego haz preguntas sobre el documento.</p>';
        return;
    }

    messages.forEach((m) => appendChatMessage(m.role, m.content));
}

window.openCity = function (evt, tabName) {
    const tabs = document.getElementsByClassName("tabcontent");
    for (let i = 0; i < tabs.length; i += 1) {
        tabs[i].style.display = "none";
    }

    const tablinks = document.getElementsByClassName("tablinks");
    for (let i = 0; i < tablinks.length; i += 1) {
        tablinks[i].classList.remove("active");
    }

    const target = document.getElementById(tabName);
    if (target) target.style.display = "block";
    if (evt && evt.currentTarget) evt.currentTarget.classList.add("active");
};

engineEl.addEventListener("change", () => {
    populateModels(engineEl.value);
});

pdfInput.addEventListener("change", () => {
    const file = pdfInput.files && pdfInput.files[0];

    if (!file) {
        if (togglePdfBtn) togglePdfBtn.disabled = true;
        if (pdfViewer) pdfViewer.removeAttribute("src");
        return;
    }

    if (file.type !== "application/pdf") {
        alert("Solo se permiten archivos PDF.");
        pdfInput.value = "";
        if (togglePdfBtn) togglePdfBtn.disabled = true;
        if (pdfViewer) pdfViewer.removeAttribute("src");
        return;
    }

    if (pdfObjectUrl) {
        URL.revokeObjectURL(pdfObjectUrl);
    }

    pdfObjectUrl = URL.createObjectURL(file);
    if (pdfViewer) pdfViewer.src = pdfObjectUrl;
});

form.addEventListener("submit", async (e) => {
    e.preventDefault();

    const engine = engineEl.value;
    const model = modelEl.value;
    const file = pdfInput.files && pdfInput.files[0];

    if (!engine || !model) {
        alert("Completa engine y modelo.");
        return;
    }

    if (!file) {
        alert("Debes seleccionar un PDF.");
        return;
    }

    submitBtn.disabled = true;
    setLoading(true, "Enviando solicitud...");
    resultEl.textContent = "Procesando...";
    outputTokenCountEl.textContent = "";

    try {
        const formdata = new FormData();
        formdata.append("file", file, file.name);

        const url =
            `api/summarize?engine=${encodeURIComponent(engine)}` +
            `&model=${encodeURIComponent(model)}`;

        const response = await fetch(url, {
            method: "POST",
            body: formdata,
            redirect: "follow"
        });

        const resultText = await response.json().catch(() => ({}));
        if (!response.ok) {
            throw new Error(resultText.error || `Error HTTP ${response.status}`);
        }

        currentDocumentId = resultText.document_id ?? null;
        if (chatStatusEl) {
            chatStatusEl.textContent = currentDocumentId
                ? "Chat listo para este documento."
                : "No se recibió document_id; el chat no estará disponible.";
        }

        if (currentDocumentId) {
            renderChatHistory(loadChatHistory(currentDocumentId));
        }

        resultEl.innerHTML = marked.parse(resultText.summary ?? "Sin contenido en la respuesta.");
        outputTokenCountEl.innerHTML =
            `<ul>
                <li>Input Tokens (aprox): ${estimateTokens(resultText.full ?? "")}</li>
                <li>Output Tokens (aprox): ${estimateTokens(resultText.summary ?? "")}</li>
            </ul>`;

        setLoading(false, "Completado");
    } catch (error) {
        setLoading(false, "Error");
        currentDocumentId = null;
        resultEl.textContent = `No se pudo completar la solicitud: ${error.message}`;
    } finally {
        submitBtn.disabled = false;
    }
});

if (chatForm) {
    chatForm.addEventListener("submit", async (e) => {
        e.preventDefault();

        if (!currentDocumentId) {
            if (chatStatusEl) chatStatusEl.textContent = "Primero genera un resumen para habilitar el chat.";
            return;
        }

        const userMessage = (chatInputEl?.value || "").trim();
        if (!userMessage) return;

        const history = loadChatHistory(currentDocumentId);
        history.push({ role: "user", content: userMessage });
        saveChatHistory(currentDocumentId, history);

        appendChatMessage("user", userMessage);
        if (chatInputEl) chatInputEl.value = "";
        if (chatSubmitBtn) chatSubmitBtn.disabled = true;
        if (chatStatusEl) chatStatusEl.textContent = "Consultando documento...";

        try {
            const engine = engineEl.value;
            const model = modelEl.value;
            const url =
                `/api/document-chat/${encodeURIComponent(currentDocumentId)}` +
                `?engine=${encodeURIComponent(engine)}&model=${encodeURIComponent(model)}`;

            const response = await fetch(url, {
                method: "POST",
                headers: { "Content-Type": "application/json; charset=utf-8" },
                body: JSON.stringify(history)
            });

            const payload = await response.json().catch(() => ({}));
            if (!response.ok) {
                throw new Error(payload.error || `Error HTTP ${response.status}`);
            }

            appendChatMessage("assistant", payload.message ?? "Sin respuesta.");
            history.push({ role: "assistant", content: payload.message ?? "Sin respuesta." });
            saveChatHistory(currentDocumentId, history);
            if (chatStatusEl) chatStatusEl.textContent = "Respuesta recibida.";
        } catch (error) {
            if (chatStatusEl) chatStatusEl.textContent = `Error: ${error.message}`;
            appendChatMessage("assistant", `No se pudo obtener respuesta: ${error.message}`);
        } finally {
            if (chatSubmitBtn) chatSubmitBtn.disabled = false;
        }
    });
}

window.addEventListener("beforeunload", () => {
    if (pdfObjectUrl) URL.revokeObjectURL(pdfObjectUrl);
});

populateModels("");
updateInputTokenCount();
setLoading(false);

const defaultTabBtn = document.querySelector(".tablinks");
if (defaultTabBtn) {
    defaultTabBtn.click();
}
