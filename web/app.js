(() => {
  const API = '';
  let token = localStorage.getItem('ripple_token');
  let myUserId = '';
  let myUsername = '';
  let currentChatId = null;
  let chats = [];
  let pollTimer = null;
  let lastMsgTimestamp = {};

  const $ = (s, p) => (p || document).querySelector(s);
  const $$ = (s, p) => [...(p || document).querySelectorAll(s)];

  async function api(method, path, body) {
    const opts = { method, headers: {} };
    if (token) opts.headers['Authorization'] = `Bearer ${token}`;
    if (body) {
      opts.headers['Content-Type'] = 'application/json';
      opts.body = JSON.stringify(body);
    }
    const res = await fetch(API + path, opts);
    const text = await res.text();
    let data;
    try { data = JSON.parse(text); } catch { data = text; }
    if (!res.ok) throw new Error(data?.error || `HTTP ${res.status}`);
    return data;
  }

  function avatarColor(id) {
    let h = 0;
    for (let i = 0; i < id.length; i++) h = (h * 31 + id.charCodeAt(i)) | 0;
    return `avatar-${Math.abs(h) % 6}`;
  }

  function initials(id) {
    return id.substring(0, 2).toUpperCase();
  }

  function formatTime(ts) {
    if (!ts) return '';
    const d = new Date(ts);
    if (isNaN(d)) return '';
    return d.toLocaleTimeString('ru-RU', { hour: '2-digit', minute: '2-digit' });
  }

  function formatDate(ts) {
    if (!ts) return '';
    const d = new Date(ts);
    if (isNaN(d)) return '';
    const today = new Date();
    if (d.toDateString() === today.toDateString()) return 'Сегодня';
    const yesterday = new Date(today);
    yesterday.setDate(yesterday.getDate() - 1);
    if (d.toDateString() === yesterday.toDateString()) return 'Вчера';
    return d.toLocaleDateString('ru-RU', { day: 'numeric', month: 'long' });
  }

  function escapeHtml(s) {
    return s.replace(/[&<>"']/g, c => ({ '&': '&amp;', '<': '&lt;', '>': '&gt;', '"': '&quot;', "'": '&#39;' }[c]));
  }

  // ─── AUTH ───

  function showAuth() {
    $('#auth-screen').style.display = '';
    $('#app').style.display = 'none';
    stopPolling();
  }

  function showApp() {
    $('#auth-screen').style.display = 'none';
    $('#app').style.display = '';
  }

  $$('.auth-tab').forEach(tab => {
    tab.addEventListener('click', () => {
      $$('.auth-tab').forEach(t => t.classList.remove('active'));
      tab.classList.add('active');
      const isRegister = tab.dataset.tab === 'register';
      $('#auth-submit-btn').textContent = isRegister ? 'Зарегистрироваться' : 'Войти';
    });
  });

  $('#auth-form').addEventListener('submit', async (e) => {
    e.preventDefault();
    const username = $('#auth-username').value.trim();
    const password = $('#auth-password').value;
    const isRegister = $('.auth-tab.active').dataset.tab === 'register';
    const errEl = $('#auth-error');
    errEl.style.display = 'none';

    try {
      if (isRegister) {
        await api('POST', '/users', { username, password });
      }
      const data = await api('POST', '/login', { username, password });
      token = data.token;
      localStorage.setItem('ripple_token', token);
      await initApp();
    } catch (err) {
      errEl.textContent = err.message;
      errEl.style.display = '';
    }
  });

  // ─── INIT ───

  async function initApp() {
    try {
      const me = await api('GET', '/me');
      myUserId = me.id;
      myUsername = me.id.substring(0, 8);
      try {
        const users = await api('GET', `/users/search?q=${encodeURIComponent(myUsername)}`);
        const meUser = users.find(u => u.id === myUserId);
        if (meUser) myUsername = meUser.username;
      } catch {}
      $('#my-username').textContent = myUsername;
      $('#my-avatar').textContent = initials(myUserId);
      $('#my-avatar').className = `avatar avatar-sm ${avatarColor(myUserId)}`;
      showApp();
      await loadChats();
      startPolling();
    } catch {
      token = null;
      localStorage.removeItem('ripple_token');
      showAuth();
    }
  }

  // ─── CHATS ───

  async function loadChats() {
    try {
      chats = await api('GET', '/chats');
      renderChatList();
    } catch {}
  }

  function renderChatList() {
    const query = ($('#chat-search').value || '').toLowerCase();
    const list = $('#chat-list');
    const filtered = chats.filter(c => {
      const name = c.id.substring(0, 8);
      return name.toLowerCase().includes(query);
    });

    if (filtered.length === 0) {
      list.innerHTML = `<div class="chat-empty">${query ? 'Ничего не найдено' : 'Нет чатов. Создайте новый!'}</div>`;
      return;
    }

    list.innerHTML = filtered.map(c => {
      const name = c.id.substring(0, 8);
      const isActive = c.id === currentChatId;
      const preview = c.last_message ? escapeHtml(c.last_message.substring(0, 50)) : 'Нет сообщений';
      const time = formatTime(c.last_msg_time);
      return `
        <div class="chat-item ${isActive ? 'active' : ''}" data-chat-id="${c.id}">
          <div class="avatar avatar-md ${avatarColor(c.id)}">${initials(c.id)}</div>
          <div class="chat-item-info">
            <div class="chat-item-top">
              <span class="chat-item-name">${escapeHtml(name)}</span>
              <span class="chat-item-time">${time}</span>
            </div>
            <div class="chat-item-bottom">
              <span class="chat-item-preview">${preview}</span>
              <span class="chat-item-members">${c.member_count} ucz.</span>
            </div>
          </div>
        </div>`;
    }).join('');

    $$('.chat-item', list).forEach(el => {
      el.addEventListener('click', () => openChat(el.dataset.chatId));
    });
  }

  // ─── CHAT VIEW ───

  async function openChat(chatId) {
    currentChatId = chatId;
    $('#empty-state').style.display = 'none';
    $('#chat-view').style.display = '';
    const name = chatId.substring(0, 8);
    $('#chat-name').textContent = name;
    $('#chat-avatar').innerHTML = `<div class="avatar avatar-md ${avatarColor(chatId)}">${initials(chatId)}</div>`;
    renderChatList();
    await loadMembers(chatId);
    await loadMessages(chatId);
    if (window.innerWidth <= 768) {
      $('#sidebar').classList.add('hidden');
    }
  }

  async function loadMembers(chatId) {
    try {
      const members = await api('GET', `/chats/${chatId}/members`);
      $('#chat-members-count').textContent = `${members.length} участник(ов)`;
    } catch {
      $('#chat-members-count').textContent = '';
    }
  }

  async function loadMessages(chatId) {
    try {
      const msgs = await api('GET', `/chats/${chatId}/messages?limit=100`);
      renderMessages(msgs);
      if (msgs.length > 0) {
        lastMsgTimestamp[chatId] = msgs[msgs.length - 1].created_at;
      }
    } catch {}
  }

  function renderMessages(msgs) {
    const container = $('#messages');
    let html = '';
    let lastDate = '';

    for (const m of msgs) {
      const date = formatDate(m.created_at);
      if (date !== lastDate) {
        html += `<div class="msg-date"><span>${date}</span></div>`;
        lastDate = date;
      }
      const isMine = m.sender_id === myUserId;
      const time = formatTime(m.created_at);
      const senderLabel = isMine ? '' : `<div class="msg-sender">${escapeHtml(m.sender_id.substring(0, 8))}</div>`;
      html += `
        <div class="msg-row ${isMine ? 'out' : 'in'}">
          <div class="msg-bubble">
            ${senderLabel}
            <span class="msg-text">${escapeHtml(m.text)}</span>
            <span class="msg-time">${time}</span>
          </div>
        </div>`;
    }

    container.innerHTML = html;
    container.scrollTop = container.scrollHeight;
  }

  // ─── SEND ───

  $('#composer').addEventListener('submit', async (e) => {
    e.preventDefault();
    const input = $('#msg-input');
    const text = input.value.trim();
    if (!text || !currentChatId) return;

    input.value = '';
    input.style.height = 'auto';

    try {
      await api('POST', `/chats/${currentChatId}/messages`, { text });
      await loadMessages(currentChatId);
      await loadChats();
    } catch {}
  });

  $('#msg-input').addEventListener('input', function () {
    this.style.height = 'auto';
    this.style.height = Math.min(this.scrollHeight, 120) + 'px';
  });

  $('#msg-input').addEventListener('keydown', (e) => {
    if (e.key === 'Enter' && !e.shiftKey) {
      e.preventDefault();
      $('#composer').dispatchEvent(new Event('submit'));
    }
  });

  // ─── NEW CHAT MODAL ───

  $('#btn-new-chat').addEventListener('click', () => {
    $('#modal-new-chat').style.display = '';
    $('#user-search').value = '';
    $('#user-search-results').innerHTML = '';
    $('#user-search').focus();
  });

  let searchDebounce;
  $('#user-search').addEventListener('input', () => {
    clearTimeout(searchDebounce);
    searchDebounce = setTimeout(() => searchUsers('user-search', 'user-search-results', createUserChat), 300);
  });

  async function createUserChat(userId) {
    try {
      const chat = await api('POST', '/chats', {});
      await api('POST', `/chats/${chat.id}/members`, { user_id: userId });
      await loadChats();
      $('#modal-new-chat').style.display = 'none';
      openChat(chat.id);
    } catch (err) {
      alert(err.message);
    }
  }

  // ─── ADD MEMBER MODAL ───

  $('#btn-add-member').addEventListener('click', () => {
    if (!currentChatId) return;
    $('#modal-add-member').style.display = '';
    $('#member-search').value = '';
    $('#member-search-results').innerHTML = '';
    $('#member-search').focus();
  });

  $('#member-search').addEventListener('input', () => {
    clearTimeout(searchDebounce);
    searchDebounce = setTimeout(() => searchUsers('member-search', 'member-search-results', addMember), 300);
  });

  async function addMember(userId) {
    try {
      await api('POST', `/chats/${currentChatId}/members`, { user_id: userId });
      await loadMembers(currentChatId);
      $('#modal-add-member').style.display = 'none';
    } catch (err) {
      alert(err.message);
    }
  }

  async function searchUsers(inputId, resultsId, onSelect) {
    const q = $(`#${inputId}`).value.trim();
    const container = $(`#${resultsId}`);
    if (!q) { container.innerHTML = ''; return; }

    try {
      const users = await api('GET', `/users/search?q=${encodeURIComponent(q)}`);
      const filtered = users.filter(u => u.id !== myUserId);
      if (filtered.length === 0) {
        container.innerHTML = '<div class="user-list-empty">Пользователи не найдены</div>';
        return;
      }
      container.innerHTML = filtered.map(u => `
        <div class="user-item" data-user-id="${u.id}">
          <div class="avatar avatar-md ${avatarColor(u.id)}">${initials(u.id)}</div>
          <div class="user-item-info">
            <div class="user-item-name">${escapeHtml(u.username)}</div>
            <div class="user-item-id">${u.id.substring(0, 12)}...</div>
          </div>
        </div>`).join('');

      $$('.user-item', container).forEach(el => {
        el.addEventListener('click', () => onSelect(el.dataset.userId));
      });
    } catch {
      container.innerHTML = '<div class="user-list-empty">Ошибка поиска</div>';
    }
  }

  // ─── MODALS CLOSE ───

  $$('.modal-close').forEach(btn => {
    btn.addEventListener('click', () => {
      const modal = $(`#${btn.dataset.close}`);
      if (modal) modal.style.display = 'none';
    });
  });

  $$('.modal-overlay').forEach(overlay => {
    overlay.addEventListener('click', (e) => {
      if (e.target === overlay) overlay.style.display = 'none';
    });
  });

  // ─── POLLING ───

  function startPolling() {
    stopPolling();
    pollTimer = setInterval(async () => {
      if (!currentChatId) return;
      try {
        const msgs = await api('GET', `/chats/${currentChatId}/messages?limit=100`);
        renderMessages(msgs);
        if (msgs.length > 0) {
          lastMsgTimestamp[currentChatId] = msgs[msgs.length - 1].created_at;
        }
      } catch {}
      try { await loadChats(); } catch {}
    }, 3000);
  }

  function stopPolling() {
    if (pollTimer) { clearInterval(pollTimer); pollTimer = null; }
  }

  // ─── NAVIGATION ───

  $('#btn-back').addEventListener('click', () => {
    $('#sidebar').classList.remove('hidden');
  });

  $('#btn-logout').addEventListener('click', () => {
    token = null;
    localStorage.removeItem('ripple_token');
    stopPolling();
    currentChatId = null;
    showAuth();
  });

  $('#chat-search').addEventListener('input', renderChatList);

  // ─── BOOT ───

  if (token) {
    initApp();
  } else {
    showAuth();
  }
})();
