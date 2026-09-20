(() => {
  "use strict";

  const state = {
    mode: "login",
    user: JSON.parse(localStorage.getItem("yiye_user") || "null"),
    accessToken: localStorage.getItem("yiye_access_token"),
    refreshToken: localStorage.getItem("yiye_refresh_token"),
    articles: [],
    currentArticle: null,
    editingId: null,
  };

  const $ = (selector) => document.querySelector(selector);
  const $$ = (selector) => [...document.querySelectorAll(selector)];
  const authView = $("#authView");
  const appView = $("#appView");
  const authForm = $("#authForm");
  const articleList = $("#articleList");
  let toastTimer;

  function setSession(payload) {
    state.user = payload.user;
    state.accessToken = payload.token.access_token;
    state.refreshToken = payload.token.refresh_token;
    localStorage.setItem("yiye_user", JSON.stringify(state.user));
    localStorage.setItem("yiye_access_token", state.accessToken);
    localStorage.setItem("yiye_refresh_token", state.refreshToken);
  }

  function clearSession() {
    state.user = null;
    state.accessToken = null;
    state.refreshToken = null;
    state.articles = [];
    localStorage.removeItem("yiye_user");
    localStorage.removeItem("yiye_access_token");
    localStorage.removeItem("yiye_refresh_token");
  }

  async function api(path, options = {}, canRefresh = true) {
    const headers = new Headers(options.headers || {});
    if (options.body) headers.set("Content-Type", "application/json");
    if (state.accessToken) headers.set("Authorization", `Bearer ${state.accessToken}`);

    const response = await fetch(`/v1${path}`, { ...options, headers });
    if (response.status === 401 && canRefresh && path !== "/token/refresh") {
      if (state.refreshToken) {
        const refreshed = await fetch("/v1/token/refresh", {
          method: "POST",
          headers: { "Content-Type": "application/json" },
          body: JSON.stringify({ refresh_token: state.refreshToken }),
        });
        if (refreshed.ok) {
          const payload = await refreshed.json();
          state.accessToken = payload.token.access_token;
          state.refreshToken = payload.token.refresh_token;
          localStorage.setItem("yiye_access_token", state.accessToken);
          localStorage.setItem("yiye_refresh_token", state.refreshToken);
          return api(path, options, false);
        }
      }
      clearSession();
      showAuth();
      throw new Error("登录已过期，请重新登录");
    }

    const payload = await response.json().catch(() => ({}));
    if (!response.ok) throw new Error(messageFor(payload.message, response.status));
    return payload;
  }

  function messageFor(message, status) {
    const messages = {
      "invalid request": "请检查填写内容",
      "invalid email or password": "邮箱或密码不正确",
      "email already exists": "这个邮箱已经注册",
      "article not found": "这篇文章不存在或已被删除",
      "title and content cannot be empty": "标题和正文都不能为空",
      "authentication service unavailable": "登录服务暂时不可用",
    };
    return messages[message] || (status >= 500 ? "服务暂时不可用，请稍后再试" : message || "请求没有完成");
  }

  function toast(message, type = "success") {
    const element = $("#toast");
    clearTimeout(toastTimer);
    element.textContent = message;
    element.className = `toast is-visible${type === "error" ? " is-error" : ""}`;
    toastTimer = setTimeout(() => { element.className = "toast"; }, 3200);
  }

  function setAuthMode(mode) {
    state.mode = mode;
    const isRegister = mode === "register";
    $$(".register-only").forEach((element) => { element.hidden = !isRegister; });
    $("#name").required = isRegister;
    $("#password").autocomplete = isRegister ? "new-password" : "current-password";
    $("#authEyebrow").textContent = isRegister ? "从今天开始" : "欢迎回来";
    $("#authTitle").textContent = isRegister ? "创建你的写作空间" : "继续写作";
    $("#authSubmitLabel").textContent = isRegister ? "创建账户" : "登录";
    $("#authSwitchText").textContent = isRegister ? "已经有账户？" : "第一次来到一页？";
    $("#authSwitch").textContent = isRegister ? "直接登录" : "创建账户";
    $("#authError").textContent = "";
  }

  function showAuth() {
    appView.hidden = true;
    authView.hidden = false;
    setAuthMode("login");
  }

  async function showApp() {
    authView.hidden = true;
    appView.hidden = false;
    $("#accountName").textContent = state.user?.name || "作者";
    $("#accountEmail").textContent = state.user?.email || "";
    $("#avatar").textContent = initial(state.user?.name);
    switchView("articles");
    await loadArticles();
  }

  async function loadArticles() {
    $("#listLoading").hidden = false;
    $("#emptyState").hidden = true;
    articleList.innerHTML = "";
    try {
      const payload = await api("/articles");
      state.articles = payload.articles || [];
      renderArticles();
    } catch (error) {
      toast(error.message, "error");
    } finally {
      $("#listLoading").hidden = true;
    }
  }

  function renderArticles(query = "") {
    const term = query.trim().toLocaleLowerCase("zh-CN");
    const articles = state.articles.filter((article) => {
      return !term || article.title.toLocaleLowerCase("zh-CN").includes(term) || article.user?.name?.toLocaleLowerCase("zh-CN").includes(term);
    });
    $("#articleCount").textContent = state.articles.length;
    articleList.innerHTML = "";
    $("#emptyState").hidden = state.articles.length !== 0 || Boolean(term);

    if (!articles.length && term) {
      articleList.innerHTML = `<div class="empty-state"><span class="empty-mark">⌕</span><h2>没有找到相关文字</h2><p>换一个关键词再试试。</p></div>`;
      return;
    }

    articles.forEach((article, index) => {
      const row = document.createElement("button");
      row.type = "button";
      row.className = "article-row";
      row.style.animationDelay = `${Math.min(index * 45, 270)}ms`;
      row.innerHTML = `
        <time class="article-date" datetime="${escapeHtml(article.created_at)}">${formatDate(article.created_at)}</time>
        <h2 class="article-title">${escapeHtml(article.title)}</h2>
        <span class="article-author"><span class="avatar">${escapeHtml(initial(article.user?.name))}</span>${escapeHtml(article.user?.name || "匿名作者")}</span>
        <span class="article-arrow" aria-hidden="true">→</span>`;
      row.addEventListener("click", () => openArticle(article.id));
      articleList.appendChild(row);
    });
  }

  async function openArticle(id) {
    try {
      const payload = await api(`/article/${id}`);
      const article = payload.article;
      state.currentArticle = article;
      $("#dialogMeta").textContent = formatLongDate(article.created_at);
      $("#dialogTitle").textContent = article.title;
      $("#dialogContent").textContent = article.content;
      $("#dialogAuthor").textContent = article.user?.name || "匿名作者";
      $("#dialogAvatar").textContent = initial(article.user?.name);
      $("#editArticleButton").hidden = article.user?.id !== state.user?.id;
      $("#articleDialog").showModal();
    } catch (error) {
      toast(error.message, "error");
    }
  }

  function switchView(view) {
    const isEditor = view === "editor";
    $("#articlesView").hidden = isEditor;
    $("#editorView").hidden = !isEditor;
    $$(".nav-item").forEach((item) => item.classList.toggle("is-active", item.dataset.view === view));
    requestAnimationFrame(positionNavLine);
    if (isEditor && !state.editingId) resetEditor();
    window.scrollTo({ top: 0, behavior: "smooth" });
  }

  function positionNavLine() {
    const active = $(".nav-item.is-active");
    const line = $(".nav-line");
    if (!active || !line) return;
    line.style.width = `${active.offsetWidth}px`;
    line.style.transform = `translateX(${active.offsetLeft}px)`;
  }

  function resetEditor() {
    state.editingId = null;
    $("#editorForm").reset();
    $("#charCount").textContent = "0";
    $("#saveStatus").textContent = "尚未发布";
    $("#publishButton").innerHTML = `发布文章 <span aria-hidden="true">↗</span>`;
    $("#deleteButton").hidden = true;
  }

  function editArticle(article) {
    state.editingId = article.id;
    $("#editorTitle").value = article.title;
    $("#editorContent").value = article.content;
    $("#charCount").textContent = countCharacters(article.content);
    $("#saveStatus").textContent = `最后更新 ${formatLongDate(article.updated_at)}`;
    $("#publishButton").innerHTML = `保存修改 <span aria-hidden="true">↗</span>`;
    $("#deleteButton").hidden = false;
    $("#articleDialog").close();
    switchView("editor");
    $("#editorTitle").focus();
  }

  async function submitArticle(event) {
    event.preventDefault();
    const title = $("#editorTitle").value.trim();
    const content = $("#editorContent").value.trim();
    if (!title || !content) {
      toast("标题和正文都不能为空", "error");
      return;
    }
    const button = $("#publishButton");
    button.disabled = true;
    try {
      await api(state.editingId ? `/article/${state.editingId}` : "/article", {
        method: state.editingId ? "PUT" : "POST",
        body: JSON.stringify({ title, content }),
      });
      toast(state.editingId ? "修改已保存" : "文章已发布");
      resetEditor();
      await loadArticles();
      switchView("articles");
    } catch (error) {
      toast(error.message, "error");
    } finally {
      button.disabled = false;
    }
  }

  async function deleteArticle() {
    if (!state.editingId || !window.confirm("确定删除这篇文章吗？此操作无法撤销。")) return;
    try {
      await api(`/article/${state.editingId}`, { method: "DELETE" });
      toast("文章已删除");
      resetEditor();
      await loadArticles();
      switchView("articles");
    } catch (error) {
      toast(error.message, "error");
    }
  }

  async function submitAuth(event) {
    event.preventDefault();
    const submit = authForm.querySelector("button[type=submit]");
    const body = {
      email: $("#email").value.trim(),
      password: $("#password").value,
    };
    if (state.mode === "register") body.name = $("#name").value.trim();
    if (!authForm.reportValidity()) return;
    submit.disabled = true;
    $("#authError").textContent = "";
    try {
      if (state.mode === "register") {
        await api("/user/register", { method: "POST", body: JSON.stringify(body) }, false);
        toast("账户已创建，请登录");
        setAuthMode("login");
        $("#password").value = "";
        $("#password").focus();
      } else {
        const payload = await api("/user/login", { method: "POST", body: JSON.stringify(body) }, false);
        setSession(payload);
        await showApp();
      }
    } catch (error) {
      $("#authError").textContent = error.message;
    } finally {
      submit.disabled = false;
    }
  }

  function formatDate(value) {
    const date = new Date(value);
    if (Number.isNaN(date.getTime())) return "—";
    return `${String(date.getMonth() + 1).padStart(2, "0")} / ${String(date.getDate()).padStart(2, "0")}\n${date.getFullYear()}`;
  }

  function formatLongDate(value) {
    const date = new Date(value);
    if (Number.isNaN(date.getTime())) return "日期未知";
    return new Intl.DateTimeFormat("zh-CN", { year: "numeric", month: "long", day: "numeric" }).format(date);
  }

  function initial(name = "") { return [...name.trim()][0] || "Y"; }
  function countCharacters(value) { return [...value.replace(/\s/g, "")].length; }
  function escapeHtml(value = "") {
    return String(value).replace(/[&<>'"]/g, (char) => ({ "&": "&amp;", "<": "&lt;", ">": "&gt;", "'": "&#39;", '"': "&quot;" })[char]);
  }

  authForm.addEventListener("submit", submitAuth);
  $("#authSwitch").addEventListener("click", () => setAuthMode(state.mode === "login" ? "register" : "login"));
  $$('[data-view]').forEach((element) => element.addEventListener("click", (event) => {
    event.preventDefault();
    if (element.hasAttribute("data-new-article") || (element.dataset.view === "editor" && !state.editingId)) resetEditor();
    switchView(element.dataset.view);
  }));
  $("#searchInput").addEventListener("input", (event) => renderArticles(event.target.value));
  $("#editorContent").addEventListener("input", (event) => { $("#charCount").textContent = countCharacters(event.target.value); });
  $("#editorForm").addEventListener("submit", submitArticle);
  $("#cancelEditButton").addEventListener("click", () => { resetEditor(); switchView("articles"); });
  $("#deleteButton").addEventListener("click", deleteArticle);
  $("#closeDialog").addEventListener("click", () => $("#articleDialog").close());
  $("#articleDialog").addEventListener("click", (event) => { if (event.target === $("#articleDialog")) $("#articleDialog").close(); });
  $("#editArticleButton").addEventListener("click", () => editArticle(state.currentArticle));
  $("#accountButton").addEventListener("click", () => {
    const popover = $("#accountPopover");
    popover.hidden = !popover.hidden;
    $("#accountButton").setAttribute("aria-expanded", String(!popover.hidden));
  });
  $("#logoutButton").addEventListener("click", async () => {
    try { await api("/user/logout"); } catch (_) { /* Local logout still succeeds. */ }
    clearSession();
    $("#accountPopover").hidden = true;
    showAuth();
    toast("已退出登录");
  });
  document.addEventListener("click", (event) => {
    if (!event.target.closest(".account-menu")) {
      $("#accountPopover").hidden = true;
      $("#accountButton").setAttribute("aria-expanded", "false");
    }
  });
  window.addEventListener("resize", positionNavLine);

  if (state.user && state.accessToken) showApp(); else showAuth();
})();
