const chats = [
  { id: "design", name: "Design team", handle: "6 участников", status: "online", description: "Место для идей, макетов и всего, что помогает сделать продукт лучше.", initials: "DT", avatar: "avatar-blue", time: "10:42", unread: 3, preview: "Nadia: новый экран уже в Figma", messages: [["Nadia", "Выложила обновленный экран профиля. Посмотрите, пожалуйста.", "10:38"], ["You", "Выглядит чисто. Особенно новый блок с активностью.", "10:40", true], ["Nadia", "Тогда сегодня отправлю в разработку ✨", "10:42"]] },
  { id: "alex", name: "Alex Morgan", handle: "был в сети недавно", status: "online", description: "Любит хороший кофе, длинные прогулки и короткие сообщения.", initials: "AM", avatar: "avatar-orange", time: "09:18", preview: "Увидимся после работы?", messages: [["Alex", "Увидимся после работы?", "09:18"]] },
  { id: "product", name: "Product updates", handle: "12 участников", status: "12 участников", description: "Новости продукта, заметки о релизах и полезные ссылки команды.", initials: "PU", avatar: "avatar-purple", time: "вчера", preview: "Новый релиз уже доступен", messages: [["Product updates", "Новый релиз уже доступен для тестирования.", "Вчера"]] },
  { id: "mira", name: "Mira Chen", handle: "была в сети 2 ч назад", status: "online", description: "Product designer · San Francisco", initials: "MC", avatar: "avatar-pink", time: "пн", preview: "Спасибо! До завтра", messages: [["Mira", "Спасибо! До завтра", "Пн"]] }
];

const state = { selected: "design", search: "" };
const root = document.querySelector(".app-shell");
const selectedChat = () => chats.find((chat) => chat.id === state.selected) || chats[0];
const escapeHtml = (value) => value.replace(/[&<>\"]/g, (character) => ({ "&": "&amp;", "<": "&lt;", ">": "&gt;", "\"": "&quot;" }[character]));

function getRoute() {
  const match = location.hash.match(/^#\/chat\/([^/]+)/);
  if (match && chats.some((chat) => chat.id === match[1])) return { name: "chat", chatId: match[1] };
  return { name: "chats" };
}

function navigate(path, replace = false) {
  const method = replace ? "replaceState" : "pushState";
  history[method]({}, "", `#${path}`);
  syncRoute();
}

function syncRoute() {
  const route = getRoute();
  if (route.name === "chat") {
    state.selected = route.chatId;
    root.classList.add("show-conversation");
  } else {
    root.classList.remove("show-conversation");
  }
  renderChatList();
  renderConversation();
}

function renderChatList() {
  const list = document.querySelector("#chat-list");
  const visible = chats.filter((chat) => `${chat.name} ${chat.preview}`.toLowerCase().includes(state.search.toLowerCase()));
  list.innerHTML = visible.map((chat) => `<button class="chat-card ${chat.id === state.selected ? "selected" : ""}" data-chat="${chat.id}">
    <span class="avatar ${chat.avatar}">${chat.initials}</span><span class="chat-copy"><h3>${chat.name}</h3><p>${chat.preview}</p></span><span class="chat-meta"><span>${chat.time}</span>${chat.unread ? `<b class="badge">${chat.unread}</b>` : ""}</span>
  </button>`).join("") || `<p class="empty-list">Ничего не найдено</p>`;
  list.querySelectorAll("[data-chat]").forEach((button) => button.addEventListener("click", () => selectChat(button.dataset.chat)));
}

function renderConversation() {
  const chat = selectedChat();
  document.querySelector("#conversation-avatar").innerHTML = `<span class="avatar ${chat.avatar}">${chat.initials}</span>`;
  document.querySelector("#conversation-name").textContent = chat.name;
  document.querySelector("#conversation-status").textContent = chat.handle;
  document.querySelector("#details-avatar").innerHTML = `<span class="avatar ${chat.avatar}">${chat.initials}</span>`;
  document.querySelector("#details-name").textContent = chat.name;
  document.querySelector("#details-handle").textContent = chat.handle;
  document.querySelector("#details-description").textContent = chat.description;
  document.querySelector("#messages").innerHTML = `<div class="date-stamp">Сегодня</div>${chat.messages.map((message) => `<div class="message-row ${message[3] ? "mine" : ""}">${message[3] ? "" : `<span class="avatar ${chat.avatar}" style="width:30px;height:30px;border-radius:10px;font-size:9px">${chat.initials}</span>`}<div class="message-bubble">${escapeHtml(message[1])}<span class="message-time">${message[2]} ${message[3] ? "✓✓" : ""}</span></div></div>`).join("")}`;
  const messages = document.querySelector("#messages");
  messages.scrollTop = messages.scrollHeight;
}

function selectChat(id) {
  navigate(`/chat/${id}`);
}

document.querySelector("#search").addEventListener("input", (event) => { state.search = event.target.value; renderChatList(); });
document.querySelector("#composer").addEventListener("submit", (event) => {
  event.preventDefault();
  const input = document.querySelector("#message-input");
  const text = input.value.trim();
  if (!text) return;
  const chat = selectedChat();
  const time = new Date().toLocaleTimeString("ru-RU", { hour: "2-digit", minute: "2-digit" });
  chat.messages.push(["You", text, time, true]);
  chat.preview = text;
  chat.time = time;
  input.value = "";
  renderChatList();
  renderConversation();
});
document.querySelector(".mobile-back").addEventListener("click", () => navigate("/chats"));

document.querySelector(".logo").addEventListener("click", (event) => {
  event.preventDefault();
  navigate("/chats");
});

window.addEventListener("hashchange", syncRoute);
window.addEventListener("popstate", syncRoute);

if (!location.hash) navigate("/chats", true);
else syncRoute();
