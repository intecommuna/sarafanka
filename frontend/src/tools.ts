import { requireUser } from './main';
import { renderParser, bindParser } from './parser';

export function renderTools() {
  if (!requireUser()) return '';
  return `
    <section class="page-heading">
      <span class="eyebrow">Инструменты</span>
      <h1>Парсер и аналитика</h1>
      <p>Сравнивайте объявления, находите точки роста и оценивайте объект по инфраструктуре.</p>
    </section>
    <div class="tool-grid">
      <a class="tool-card" href="#/tools">
        <span class="tool-icon">🔎</span>
        <strong>Парсер сделок</strong>
        <small>Поиск и фильтрация по рынку</small>
      </a>
      <a class="tool-card" href="#/tools/analytics">
        <span class="tool-icon">📊</span>
        <strong>Аналитика объекта</strong>
        <small>Инфраструктура, курсы и ипотека</small>
      </a>
      <div class="tool-card soon">
        <span class="tool-icon">🧠</span>
        <strong>AI-предсказания</strong>
        <small>Скоро</small>
      </div>
    </div>
    ${renderParser()}
  `;
}

export function bindTools() {
  bindParser();
}
