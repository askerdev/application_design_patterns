package web

// ganttHTML — самодостаточная страница с диаграммой Ганта и сводной аналитикой
// на Mermaid (gantt + pie + xychart).
const ganttHTML = `<!doctype html>
<html lang="ru">
<head>
<meta charset="utf-8">
<title>TaskFlow — Gantt</title>
<script type="module">
  import mermaid from "https://cdn.jsdelivr.net/npm/mermaid@10.9.1/dist/mermaid.esm.min.mjs";
  mermaid.initialize({
    startOnLoad: false,
    theme: "dark",
    gantt: {
      barHeight: 22,
      barGap: 6,
      topPadding: 50,
      leftPadding: 220,
      gridLineStartPadding: 35,
      fontSize: 13,
      sectionFontSize: 14,
      numberSectionStyles: 4,
    },
    pie: { textPosition: 0.6 },
    securityLevel: "loose",
  });
  window.__mermaid = mermaid;
</script>
<style>
  body { font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", sans-serif; margin: 0; padding: 20px; background:#0f1419; color:#e6edf3; }
  h1 { margin-top:0; color:#7aa2f7; }
  h2 { color:#7aa2f7; margin-top:32px; margin-bottom:12px; font-size:18px; }
  .meta { color:#8b949e; margin-bottom:10px; font-size:14px; }
  .controls { margin-bottom:16px; }
  select { background:#1f2937; color:#e6edf3; border:1px solid #30363d; padding:6px 10px; border-radius:6px; font-size: 14px; }
  .chart-box { background:#161b22; border:1px solid #30363d; border-radius:8px; padding:16px; overflow-x:auto; }
  .empty { color:#8b949e; font-style: italic; padding: 20px; }
  .warn { background:#2d1b00; color:#f0a020; padding:8px 12px; border-radius:6px; margin-bottom:12px; font-size:13px; }
  .err { background:#3a0e0e; color:#f85149; padding:8px 12px; border-radius:6px; margin-bottom:12px; font-size:13px; white-space:pre-wrap; }
  table { width:100%; margin-top:20px; border-collapse: collapse; font-size:14px; }
  th, td { text-align:left; padding:8px; border-bottom:1px solid #30363d; }
  th { color:#8b949e; font-weight:500; }
  .pri-HIGH { color:#f85149; }
  .pri-MEDIUM { color:#d29922; }
  .pri-LOW { color:#3fb950; }
  .status-DONE { color:#8b949e; text-decoration:line-through; }
  #mermaid-container svg { max-width: none; }
  .charts-row { display: grid; grid-template-columns: 1fr 1fr; gap: 16px; margin-top: 16px; }
  @media (max-width: 1100px) { .charts-row { grid-template-columns: 1fr; } }
  .stats-cards { display: grid; grid-template-columns: repeat(4, 1fr); gap: 12px; margin-top: 16px; }
  @media (max-width: 800px) { .stats-cards { grid-template-columns: repeat(2, 1fr); } }
  .card { background:#161b22; border:1px solid #30363d; border-radius:8px; padding:14px 16px; }
  .card .label { font-size: 12px; color:#8b949e; text-transform: uppercase; letter-spacing: 0.5px; }
  .card .value { font-size: 28px; font-weight: 600; margin-top: 6px; }
  .value.bad { color:#f85149; }
  .value.warn { color:#d29922; }
  .value.good { color:#3fb950; }
  .value.muted { color:#8b949e; }
</style>
</head>
<body>
  <h1>📊 Диаграмма Ганта</h1>

  <div class="controls">
    <label>Проект:
      <select id="projectSelect" onchange="onProjectChange()">
        {{range .Projects}}
          <option value="{{.ID}}">{{.Name}}</option>
        {{end}}
      </select>
    </label>
  </div>

  <div id="meta" class="meta"></div>
  <div id="warn" class="warn" style="display:none"></div>
  <div id="err" class="err" style="display:none"></div>

  <div class="chart-box">
    <div id="mermaid-container" class="empty">Загрузка...</div>
  </div>

  <h2>📈 Сводка</h2>
  <div class="stats-cards" id="statsCards"></div>

  <div class="charts-row">
    <div>
      <h2>🥧 Доля длительности задач</h2>
      <div class="chart-box">
        <div id="pie-container" class="empty">—</div>
      </div>
    </div>
    <div>
      <h2>📅 Загрузка по дням</h2>
      <div class="chart-box">
        <div id="bar-container" class="empty">—</div>
      </div>
    </div>
  </div>

  <h2>📋 Задачи</h2>
  <table id="taskTable">
    <thead>
      <tr><th>#</th><th>Задача</th><th>Приоритет</th><th>SP</th><th>Статус</th><th>Старт</th><th>Финиш</th><th>Заметка</th></tr>
    </thead>
    <tbody></tbody>
  </table>

<script>
const initialProjectID = {{.ProjectID}};

function fmtDate(s) {
  return new Date(s).toISOString().slice(0, 10);
}

function escapeMermaid(s) {
  return String(s).replace(/:/g, "·").replace(/#/g, "№").replace(/\n/g, " ").trim();
}

function escapeHTML(s) {
  return String(s).replace(/[&<>"]/g, c => ({"&":"&amp;","<":"&lt;",">":"&gt;","\"":"&quot;"}[c]));
}

function durationDays(item) {
  // целое число дней (минимум 1) — длительность задачи в днях.
  const start = new Date(item.start);
  const end = new Date(item.end);
  const ms = end - start;
  const days = Math.max(1, Math.round(ms / (24 * 3600 * 1000)));
  return days;
}

function buildMermaid(plan) {
  const projName = plan.project ? plan.project.Name : "Project";
  const items = plan.items || [];
  if (items.length === 0) return null;

  const groups = { HIGH: [], MEDIUM: [], LOW: [] };
  for (const it of items) {
    (groups[it.priority] || groups.LOW).push(it);
  }

  let src = "gantt\n";
  src += "    dateFormat YYYY-MM-DD\n";
  src += "    axisFormat %d %b\n";
  src += "    title " + escapeMermaid(projName) + "\n";
  src += "    excludes weekends\n";

  const sectionOrder = [
    ["HIGH",   "🔥 HIGH"],
    ["MEDIUM", "⚡ MEDIUM"],
    ["LOW",    "🟢 LOW"],
  ];

  for (const [key, title] of sectionOrder) {
    const list = groups[key];
    if (!list || list.length === 0) continue;
    src += "    section " + title + "\n";
    for (const it of list) {
      const start = fmtDate(it.start);
      const end = fmtDate(it.end);
      const t = escapeMermaid(it.title) + " (#" + it.task_id + ")";
      let status = "";
      if (it.status === "DONE") status = "done, ";
      else if (it.status === "IN_PROGRESS") status = "active, ";
      else if (it.note && /OVERDUE|дедлайн/i.test(it.note)) status = "crit, ";
      src += "    " + t + " :" + status + "t" + it.task_id + ", " + start + ", " + end + "\n";
    }
  }
  return src;
}

// buildPie — Pie chart: доля длительности каждой задачи относительно суммарного срока.
// Если задач много (>10), мелкие сливаются в "Прочие" чтобы диаграмма читалась.
function buildPie(items) {
  if (!items || items.length === 0) return null;

  const enriched = items.map(it => ({
    title: it.title,
    id: it.task_id,
    days: durationDays(it),
  }));
  const totalDays = enriched.reduce((s, x) => s + x.days, 0);
  if (totalDays === 0) return null;

  // Сортируем по убыванию длительности.
  enriched.sort((a, b) => b.days - a.days);

  const TOP = 10;
  let visible = enriched;
  let otherDays = 0;
  if (enriched.length > TOP) {
    visible = enriched.slice(0, TOP);
    otherDays = enriched.slice(TOP).reduce((s, x) => s + x.days, 0);
  }

  let src = "pie showData\n";
  src += '    title Распределение времени по задачам (всего ' + totalDays + ' дней)\n';
  for (const e of visible) {
    const pct = ((e.days / totalDays) * 100).toFixed(1);
    // Mermaid pie: "Label" : value
    const label = escapeMermaid(e.title) + " (#" + e.id + ", " + pct + "%)";
    src += '    "' + label.replace(/"/g, "'") + '" : ' + e.days + "\n";
  }
  if (otherDays > 0) {
    src += '    "Прочие задачи (' + (enriched.length - TOP) + ')" : ' + otherDays + "\n";
  }
  return src;
}

// buildDailyBar — bar chart: сколько задач активно в каждый день периода.
// Используем Mermaid xychart-beta (доступен с 10.6+).
function buildDailyBar(items) {
  if (!items || items.length === 0) return null;

  // Найдём минимальную и максимальную даты.
  let minDate = null, maxDate = null;
  for (const it of items) {
    const s = new Date(it.start);
    const e = new Date(it.end);
    s.setHours(0, 0, 0, 0);
    e.setHours(0, 0, 0, 0);
    if (!minDate || s < minDate) minDate = s;
    if (!maxDate || e > maxDate) maxDate = e;
  }
  if (!minDate || !maxDate) return null;

  // Строим массив дней.
  const days = [];
  const counts = [];
  const labels = [];
  const oneDay = 24 * 3600 * 1000;
  const totalDays = Math.round((maxDate - minDate) / oneDay) + 1;

  // Если период слишком длинный — Mermaid xychart станет нечитаем.
  // Ограничим хвост 30 днями (часто проекты гораздо длиннее).
  const MAX_DAYS = 30;
  const cutEnd = totalDays > MAX_DAYS
    ? new Date(minDate.getTime() + (MAX_DAYS - 1) * oneDay)
    : maxDate;

  for (let d = new Date(minDate); d <= cutEnd; d = new Date(d.getTime() + oneDay)) {
    let cnt = 0;
    for (const it of items) {
      const s = new Date(it.start); s.setHours(0, 0, 0, 0);
      const e = new Date(it.end);   e.setHours(0, 0, 0, 0);
      if (d >= s && d <= e) cnt++;
    }
    days.push(new Date(d));
    counts.push(cnt);
    // Краткий формат даты для оси: "DD.MM" или "DD MMM"
    const dd = String(d.getDate()).padStart(2, "0");
    const mm = String(d.getMonth() + 1).padStart(2, "0");
    labels.push(dd + "." + mm);
  }

  if (counts.length === 0) return null;
  const maxCnt = Math.max(...counts, 1);

  let src = "xychart-beta\n";
  src += '    title "Количество активных задач по дням' + (totalDays > MAX_DAYS ? " (первые " + MAX_DAYS + " дней)" : "") + '"\n';
  src += '    x-axis [' + labels.map(l => '"' + l + '"').join(", ") + ']\n';
  src += '    y-axis "Задач" 0 --> ' + (maxCnt + 1) + "\n";
  src += '    bar [' + counts.join(", ") + ']\n';
  return src;
}

// computeStats — собирает агрегаты для карточек: всего задач, done %, overdue %, % SP overdue.
function computeStats(plan) {
  const items = plan.items || [];
  const total = items.length;
  let done = 0, inProg = 0, todo = 0;
  let overdueTasks = 0;
  let totalSP = 0, overdueSP = 0;
  let projDeadline = null;
  if (plan.project && plan.project.DueDate) {
    projDeadline = new Date(plan.project.DueDate);
  }

  const today = new Date();
  today.setHours(0, 0, 0, 0);

  for (const it of items) {
    if (it.status === "DONE") done++;
    else if (it.status === "IN_PROGRESS") inProg++;
    else todo++;

    totalSP += it.story_points || 0;

    // Задача считается просроченной, если её планируемый end > project deadline,
    // либо в note есть слово OVERDUE (от fallback / LLM), либо end < today и статус != DONE.
    let isOverdue = false;
    const end = new Date(it.end); end.setHours(0, 0, 0, 0);
    if (it.note && /OVERDUE|просроч|дедлайн/i.test(it.note)) isOverdue = true;
    if (projDeadline && end > projDeadline) isOverdue = true;
    if (end < today && it.status !== "DONE") isOverdue = true;

    if (isOverdue) {
      overdueTasks++;
      overdueSP += it.story_points || 0;
    }
  }

  const donePct = total > 0 ? Math.round((done / total) * 100) : 0;
  const overduePct = total > 0 ? Math.round((overdueTasks / total) * 100) : 0;
  const overdueSPPct = totalSP > 0 ? Math.round((overdueSP / totalSP) * 100) : 0;

  return {
    total, done, inProg, todo,
    overdueTasks, overduePct,
    totalSP, overdueSP, overdueSPPct,
    donePct,
  };
}

function renderStatsCards(stats) {
  const cards = document.getElementById("statsCards");
  const overdueClass = stats.overduePct >= 30 ? "bad" : stats.overduePct >= 10 ? "warn" : "good";
  const spClass = stats.overdueSPPct >= 30 ? "bad" : stats.overdueSPPct >= 10 ? "warn" : "good";
  cards.innerHTML =
    card("Всего задач", stats.total, "muted") +
    card("Сделано", stats.donePct + "%", "good", stats.done + " из " + stats.total) +
    card("Просрочено задач", stats.overduePct + "%", overdueClass, stats.overdueTasks + " шт.") +
    card("Просрочено SP", stats.overdueSPPct + "%", spClass, stats.overdueSP + " из " + stats.totalSP + " SP");
}

function card(label, value, cls, sub) {
  return '<div class="card">' +
    '<div class="label">' + escapeHTML(label) + '</div>' +
    '<div class="value ' + cls + '">' + escapeHTML(String(value)) + '</div>' +
    (sub ? '<div class="label" style="margin-top:6px;text-transform:none">' + escapeHTML(sub) + '</div>' : '') +
    '</div>';
}

function fillTable(items) {
  const tbody = document.querySelector("#taskTable tbody");
  tbody.innerHTML = "";
  for (const it of items) {
    const tr = document.createElement("tr");
    tr.innerHTML =
      "<td>#" + it.task_id + "</td>" +
      "<td class=\"status-" + it.status + "\">" + escapeHTML(it.title) + "</td>" +
      "<td class=\"pri-" + it.priority + "\">" + it.priority + "</td>" +
      "<td>" + it.story_points + "</td>" +
      "<td>" + it.status + "</td>" +
      "<td>" + fmtDate(it.start) + "</td>" +
      "<td>" + fmtDate(it.end) + "</td>" +
      "<td>" + escapeHTML(it.note || "") + "</td>";
    tbody.appendChild(tr);
  }
}

async function renderMermaidInto(containerID, src, errCtx) {
  const container = document.getElementById(containerID);
  if (!src) {
    container.className = "empty";
    container.textContent = "Недостаточно данных для построения.";
    return;
  }
  try {
    const mermaid = window.__mermaid;
    if (!mermaid) {
      setTimeout(() => renderMermaidInto(containerID, src, errCtx), 200);
      return;
    }
    const id = containerID + "-svg-" + Date.now();
    const { svg } = await mermaid.render(id, src);
    container.className = "";
    container.innerHTML = svg;
  } catch (e) {
    container.className = "empty";
    container.textContent = "Не удалось отрендерить (" + errCtx + ").";
    const errEl = document.getElementById("err");
    errEl.textContent = errCtx + " render error:\n" + (e && e.message ? e.message : String(e)) + "\n\nИсточник:\n" + src;
    errEl.style.display = "block";
  }
}

async function loadPlan(projectID) {
  const meta = document.getElementById("meta");
  const warnEl = document.getElementById("warn");
  const errEl = document.getElementById("err");
  const container = document.getElementById("mermaid-container");
  warnEl.style.display = "none";
  errEl.style.display = "none";
  meta.textContent = "Загрузка...";
  container.className = "empty";
  container.textContent = "Загрузка...";

  let plan;
  try {
    const resp = await fetch("/api/gantt?project_id=" + projectID);
    if (!resp.ok) {
      const text = await resp.text();
      meta.textContent = "Ошибка загрузки";
      errEl.textContent = "HTTP " + resp.status + ": " + text;
      errEl.style.display = "block";
      return;
    }
    plan = await resp.json();
  } catch (e) {
    meta.textContent = "Ошибка сети";
    errEl.textContent = String(e);
    errEl.style.display = "block";
    return;
  }

  const projName = plan.project ? plan.project.Name : "—";
  const items = plan.items || [];
  const used = plan.used_llm ? "🤖 LLM" : "⚙️ fallback";
  meta.textContent = "Проект: " + projName + " · задач: " + items.length + " · сгенерировано: " + used;

  if (plan.warning) {
    warnEl.textContent = "⚠ " + plan.warning;
    warnEl.style.display = "block";
  }

  fillTable(items);
  renderStatsCards(computeStats(plan));

  // Параллельно рендерим три диаграммы.
  await Promise.all([
    renderMermaidInto("mermaid-container", buildMermaid(plan), "Gantt"),
    renderMermaidInto("pie-container", buildPie(items), "Pie"),
    renderMermaidInto("bar-container", buildDailyBar(items), "Bar"),
  ]);
}

function onProjectChange() {
  const sel = document.getElementById("projectSelect");
  const pid = sel.value;
  history.replaceState(null, "", "/?project_id=" + pid);
  loadPlan(pid);
}

(function init() {
  const sel = document.getElementById("projectSelect");
  if (initialProjectID && sel.querySelector('option[value="' + initialProjectID + '"]')) {
    sel.value = initialProjectID;
  }
  if (sel && sel.value) {
    loadPlan(sel.value);
  } else {
    document.getElementById("meta").textContent = "Нет проектов.";
    document.getElementById("mermaid-container").textContent = "Нет проектов.";
  }
})();
</script>
</body>
</html>`
