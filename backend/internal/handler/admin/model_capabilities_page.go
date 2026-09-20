package admin

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// ModelCapabilitiesPage 提供「模型对外暴露」配置页。
// 页面本身是静态资源（无鉴权），数据由 /api/v1/admin/settings/model-capabilities
// 与分组接口提供，复用后台登录态（localStorage.auth_token）。
func (h *SettingHandler) ModelCapabilitiesPage(c *gin.Context) {
	c.Data(http.StatusOK, "text/html; charset=utf-8", []byte(modelCapabilitiesPageHTML))
}

// ModelCapabilitiesPageScript 提供页面脚本（CSP 只允许同源脚本，故单独成文件）。
func (h *SettingHandler) ModelCapabilitiesPageScript(c *gin.Context) {
	c.Data(http.StatusOK, "application/javascript; charset=utf-8", []byte(modelCapabilitiesPageJS))
}

const modelCapabilitiesPageHTML = `<!doctype html>
<html lang="zh-CN">
<head>
<meta charset="utf-8">
<meta name="viewport" content="width=device-width,initial-scale=1">
<title>模型对外暴露 · Sub2API</title>
<style>
  body{font:14px/1.5 -apple-system,BlinkMacSystemFont,"Segoe UI",Roboto,"PingFang SC","Microsoft YaHei",sans-serif;margin:0;background:#0f1115;color:#e6e8eb}
  header{padding:14px 20px;border-bottom:1px solid #262b33;display:flex;gap:14px;align-items:center;flex-wrap:wrap}
  h1{font-size:16px;margin:0 8px 0 0}
  main{padding:14px 20px}
  table{border-collapse:collapse;width:100%}
  th,td{border-bottom:1px solid #262b33;padding:6px 8px;text-align:left;vertical-align:middle}
  th{color:#9aa4b2;font-weight:600;position:sticky;top:0;background:#0f1115}
  code{font-family:ui-monospace,SFMono-Regular,Menlo,monospace}
  input[type=text],input[type=number],select{background:#161a21;color:#e6e8eb;border:1px solid #2c333d;border-radius:6px;padding:4px 6px;font:inherit}
  input[type=number]{width:110px}
  .levels label{display:inline-flex;align-items:center;gap:3px;margin-right:6px;font-size:12px;color:#c3cad4;white-space:nowrap}
  button{background:#2f6feb;color:#fff;border:0;border-radius:6px;padding:6px 12px;font:inherit;cursor:pointer}
  button.secondary{background:#2c333d}
  button.link{background:none;color:#7f8b99;padding:2px 4px}
  .badge{font-size:12px;border-radius:999px;padding:2px 8px;background:#1f2937;color:#9aa4b2}
  .badge.settings{background:#16351f;color:#7ee2a8}
  .muted{color:#7f8b99}
  .status{margin-left:auto;color:#7ee2a8}
  .status.error{color:#f5737b}
</style>
</head>
<body>
<header>
  <h1>模型对外暴露</h1>
  <label>分组 <select id="group"></select></label>
  <label class="muted"><input type="checkbox" id="allowlist-enabled"> 启用白名单</label>
  <span class="badge" id="source">—</span>
  <span class="status" id="status"></span>
</header>
<main>
  <p class="muted" id="hint">勾选「暴露」= 该模型出现在 /v1/models；白名单关闭时列出的是账号映射的全部模型。context_length / 思考等级会写进 /v1/models，供 CC Switch 等客户端读取。</p>
  <table>
    <thead><tr><th>模型 ID</th><th>暴露</th><th>context_length</th><th>思考等级</th><th>默认等级</th><th></th></tr></thead>
    <tbody id="rows"></tbody>
  </table>
  <p><input type="text" id="new-model" placeholder="新增模型 id" size="28"> <button class="secondary" id="add">添加</button>
     <button id="save-capabilities">保存能力</button>
     <button class="secondary" id="save-allowlist">保存白名单</button></p>
  <p class="muted" id="note"></p>
</main>
<script src="app.js"></script>
</body>
</html>
`

const modelCapabilitiesPageJS = `(function () {
  'use strict';
  var state = { levels: [], groups: [], group: null, candidates: [], caps: {}, exposed: [], enabled: false };

  function token() { return localStorage.getItem('auth_token') || ''; }

  function api(path, options) {
    options = options || {};
    var headers = { 'Authorization': 'Bearer ' + token() };
    if (options.body) { headers['Content-Type'] = 'application/json'; }
    return fetch(path, { method: options.method || 'GET', headers: headers, body: options.body })
      .then(function (res) {
        if (res.status === 401) { throw new Error('未登录后台：请先在同一浏览器登录管理后台'); }
        return res.json().catch(function () { return null; }).then(function (body) {
          if (!res.ok) {
            var message = body && (body.message || (body.error && body.error.message));
            throw new Error(message || ('HTTP ' + res.status));
          }
          return body ? body.data : null;
        });
      });
  }

  function setStatus(text, isError) {
    var el = document.getElementById('status');
    el.textContent = text || '';
    el.className = isError ? 'status error' : 'status';
    if (text) { window.setTimeout(function () { if (el.textContent === text) { el.textContent = ''; } }, 4000); }
  }

  function capOf(id) { return state.caps[id] || (state.caps[id] = {}); }

  function levelEnabled(id, level) {
    var levels = capOf(id).reasoning_levels || [];
    return levels.indexOf(level) >= 0;
  }

  function rowIDs() {
    var seen = {}, ids = [];
    function push(id) { id = (id || '').trim(); if (id && !seen[id]) { seen[id] = 1; ids.push(id); } }
    state.candidates.forEach(push);
    Object.keys(state.caps).forEach(push);
    state.exposed.forEach(push);
    return ids;
  }

  function render() {
    var levels = state.levels;
    var body = document.getElementById('rows');
    body.innerHTML = '';
    rowIDs().forEach(function (id) {
      var cap = capOf(id);
      var tr = document.createElement('tr');

      var tdID = document.createElement('td');
      var code = document.createElement('code');
      code.textContent = id;
      tdID.appendChild(code);
      tr.appendChild(tdID);

      var tdExposed = document.createElement('td');
      var exposed = document.createElement('input');
      exposed.type = 'checkbox';
      exposed.checked = state.exposed.indexOf(id) >= 0;
      exposed.onchange = function () {
        var index = state.exposed.indexOf(id);
        if (exposed.checked && index < 0) { state.exposed.push(id); }
        if (!exposed.checked && index >= 0) { state.exposed.splice(index, 1); }
      };
      tdExposed.appendChild(exposed);
      tr.appendChild(tdExposed);

      var tdCtx = document.createElement('td');
      var ctx = document.createElement('input');
      ctx.type = 'number';
      ctx.min = '0';
      ctx.step = '1000';
      ctx.value = cap.context_length || '';
      ctx.oninput = function () { cap.context_length = parseInt(ctx.value, 10) || 0; };
      tdCtx.appendChild(ctx);
      tr.appendChild(tdCtx);

      var tdLevels = document.createElement('td');
      tdLevels.className = 'levels';
      levels.forEach(function (level) {
        var label = document.createElement('label');
        var box = document.createElement('input');
        box.type = 'checkbox';
        box.checked = levelEnabled(id, level);
        box.onchange = function () {
          var list = cap.reasoning_levels || (cap.reasoning_levels = []);
          var index = list.indexOf(level);
          if (box.checked && index < 0) { list.push(level); }
          if (!box.checked && index >= 0) { list.splice(index, 1); }
          if (cap.default_reasoning_level && list.indexOf(cap.default_reasoning_level) < 0) { cap.default_reasoning_level = ''; }
          render();
        };
        label.appendChild(box);
        label.appendChild(document.createTextNode(level));
        tdLevels.appendChild(label);
      });
      tr.appendChild(tdLevels);

      var tdDefault = document.createElement('td');
      var select = document.createElement('select');
      var empty = document.createElement('option');
      empty.value = '';
      empty.textContent = '—';
      select.appendChild(empty);
      (cap.reasoning_levels || []).forEach(function (level) {
        var option = document.createElement('option');
        option.value = level;
        option.textContent = level;
        select.appendChild(option);
      });
      select.value = cap.default_reasoning_level || '';
      select.onchange = function () { cap.default_reasoning_level = select.value; };
      tdDefault.appendChild(select);
      tr.appendChild(tdDefault);

      var tdRemove = document.createElement('td');
      var remove = document.createElement('button');
      remove.className = 'link';
      remove.textContent = '移除';
      remove.onclick = function () {
        delete state.caps[id];
        state.candidates = state.candidates.filter(function (item) { return item !== id; });
        state.exposed = state.exposed.filter(function (item) { return item !== id; });
        render();
      };
      tdRemove.appendChild(remove);
      tr.appendChild(tdRemove);

      body.appendChild(tr);
    });
  }

  function loadCapabilities() {
    return api('/api/v1/admin/settings/model-capabilities').then(function (data) {
      state.levels = (data && data.reasoning_levels) || [];
      state.caps = {};
      ((data && data.models) || []).forEach(function (item) {
        state.caps[item.id] = {
          context_length: item.context_length || 0,
          reasoning_levels: item.reasoning_levels || [],
          default_reasoning_level: item.default_reasoning_level || ''
        };
      });
      var badge = document.getElementById('source');
      badge.textContent = data && data.source === 'settings' ? '来源：后台设置' : '来源：config.yaml';
      badge.className = 'badge' + (data && data.source === 'settings' ? ' settings' : '');
    });
  }

  function loadGroups() {
    return api('/api/v1/admin/groups/all').then(function (groups) {
      state.groups = groups || [];
      var select = document.getElementById('group');
      select.innerHTML = '';
      state.groups.forEach(function (group) {
        var option = document.createElement('option');
        option.value = String(group.id);
        option.textContent = group.name + '（' + group.platform + '）';
        select.appendChild(option);
      });
      select.onchange = function () { selectGroup(select.value); };
      if (state.groups.length) { return selectGroup(String(state.groups[0].id)); }
    });
  }

  function selectGroup(groupID) {
    state.group = state.groups.filter(function (item) { return String(item.id) === String(groupID); })[0] || null;
    var allowlist = (state.group && state.group.model_allowlist) || {};
    state.enabled = !!allowlist.enabled;
    state.exposed = (allowlist.models || []).slice();
    document.getElementById('allowlist-enabled').checked = state.enabled;
    document.getElementById('group').value = groupID;
    document.getElementById('note').textContent = state.enabled
      ? '白名单已启用：只有勾选的模型会出现在 /v1/models。'
      : '白名单未启用：/v1/models 暴露该分组账号映射的全部模型，勾选内容在启用后生效。';
    return api('/api/v1/admin/groups/' + encodeURIComponent(groupID) + '/model-allowlist-candidates').then(function (data) {
      state.candidates = (data && data.models) || [];
      render();
    });
  }

  function saveCapabilities() {
    var models = Object.keys(state.caps).map(function (id) {
      var cap = state.caps[id];
      var item = { id: id };
      if (cap.context_length) { item.context_length = cap.context_length; }
      if ((cap.reasoning_levels || []).length) { item.reasoning_levels = cap.reasoning_levels; }
      if (cap.default_reasoning_level) { item.default_reasoning_level = cap.default_reasoning_level; }
      return item;
    });
    return api('/api/v1/admin/settings/model-capabilities', { method: 'PUT', body: JSON.stringify({ models: models }) })
      .then(function () { return loadCapabilities(); })
      .then(function () { setStatus('能力已保存'); })
      .catch(function (error) { setStatus(error.message, true); });
  }

  function saveAllowlist() {
    if (!state.group) { return Promise.resolve(); }
    var payload = { model_allowlist: { enabled: document.getElementById('allowlist-enabled').checked, models: state.exposed.slice() } };
    return api('/api/v1/admin/groups/' + encodeURIComponent(state.group.id), { method: 'PUT', body: JSON.stringify(payload) })
      .then(function () { return loadGroups(); })
      .then(function () { setStatus('白名单已保存'); })
      .catch(function (error) { setStatus(error.message, true); });
  }

  document.getElementById('save-capabilities').onclick = saveCapabilities;
  document.getElementById('save-allowlist').onclick = saveAllowlist;
  document.getElementById('add').onclick = function () {
    var input = document.getElementById('new-model');
    var id = input.value.trim();
    if (!id) { return; }
    capOf(id);
    if (state.exposed.indexOf(id) < 0) { state.exposed.push(id); }
    input.value = '';
    render();
  };

  loadCapabilities().then(loadGroups).catch(function (error) { setStatus(error.message, true); });
})();
`
