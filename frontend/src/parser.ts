import { getParserResults } from './api';
import { escapeHtml } from './main';

const sourceColors: Record<string, string> = {
  avito: '#9B59B6',
  cian: '#4A90E2',
  domclick: '#E74C3C'
};

const cityOptions = ['Москва', 'Санкт-Петербург', 'Казань', 'Екатеринбург', 'Новосибирск', 'Махачкала', 'Южно-Сухокумск'];

export function renderParser() {
  return `
    <section class="tool-panel">
      <div class="panel-header">
        <div>
          <span class="eyebrow">Парсер</span>
          <h2>Реальные предложения в одном окне</h2>
        </div>
      </div>
      <form id="parser-form" class="parser-form">
        <div class="parser-row">
          <label>
            <span>Город</span>
            <input name="city" value="Москва" placeholder="Москва" list="cities-list" />
            <datalist id="cities-list">${cityOptions.map((city) => `<option value="${escapeHtml(city)}"></option>`).join('')}</datalist>
          </label>
          <label>
            <span>Район</span>
            <input name="district" value="" placeholder="Любой" />
          </label>
          <label>
            <span>Комнат</span>
            <input name="rooms" type="number" min="1" max="5" value="2" />
          </label>
          <label>
            <span>Мин. цена</span>
            <input name="price_min" type="number" value="5000000" />
          </label>
          <label>
            <span>Макс. цена</span>
            <input name="price_max" type="number" value="20000000" />
          </label>
          <label>
            <span>Источник</span>
            <select name="source">
              <option value="">Все</option>
              <option value="avito">Авито</option>
              <option value="cian">Циан</option>
              <option value="domclick">Домклик</option>
            </select>
          </label>
        </div>
        <div class="parser-actions">
          <button class="button primary" type="submit">Найти</button>
          <select name="sort">
            <option value="price_asc">Сначала дешёвые</option>
            <option value="price_desc">Сначала дорогие</option>
            <option value="rooms_desc">Больше комнат</option>
          </select>
        </div>
      </form>
      <div id="parser-results" class="parser-results"></div>
    </section>
  `;
}

export function bindParser() {
  const form = document.querySelector<HTMLFormElement>('#parser-form');
  if (!form) return;
  const results = document.querySelector<HTMLElement>('#parser-results');
  if (!results) return;

  const load = async () => {
    results.innerHTML = skeletons();
    const formData = new FormData(form);
    const params = new URLSearchParams();
    const city = formData.get('city')?.toString().trim() || '';
    const district = formData.get('district')?.toString().trim() || '';
    const rooms = formData.get('rooms')?.toString().trim() || '';
    const priceMin = formData.get('price_min')?.toString().trim() || '';
    const priceMax = formData.get('price_max')?.toString().trim() || '';
    const source = formData.get('source')?.toString().trim() || '';
    const sort = formData.get('sort')?.toString().trim() || 'price_asc';
    if (city) params.set('city', city);
    if (district) params.set('district', district);
    if (rooms) params.set('rooms', rooms);
    if (priceMin) params.set('price_min', priceMin);
    if (priceMax) params.set('price_max', priceMax);
    if (source) params.set('source', source);
    if (sort) params.set('sort', sort);
    try {
      const response = await getParserResults(Object.fromEntries(params.entries()));
      results.innerHTML = response.items.length ? response.items.map(item => {
        const sourceLabel = item.source === 'avito' ? 'Авито' : item.source === 'cian' ? 'Циан' : 'Домклик';
        const sourceUrl = item.search_url || `https://www.${item.source === 'domclick' ? 'domclick.ru' : item.source === 'avito' ? 'avito.ru' : 'cian.ru'}`;
        return `
          <article class="parser-card">
            <div class="parser-top">
              <span class="source-badge" style="background:${sourceColors[item.source] || '#4A90E2'};">${escapeHtml(item.source)}</span>
              <span class="item-price">${Number(item.price).toLocaleString('ru-RU')} ₽</span>
            </div>
            <h3>${escapeHtml(item.title)}</h3>
            <p>${escapeHtml(item.city)} · ${escapeHtml(item.district)}</p>
            <div class="meta-row">
              <span>${item.rooms} комн.</span>
              <span>${escapeHtml(item.address)}</span>
            </div>
            <div class="parser-actions">
              <a class="button primary" href="${sourceUrl}" target="_blank" rel="noreferrer">Поиск на ${sourceLabel}</a>
              <a class="button alt" href="#/tools/analytics?address=${encodeURIComponent(item.address)}">📊 Аналитика адреса</a>
            </div>
          </article>
        `;
      }).join('') : '<div class="empty">По вашему запросу нет подходящих вариантов.</div>';
    } catch (error) {
      results.innerHTML = `<div class="error">${escapeHtml((error as Error).message)}</div>`;
    }
  };

  form.addEventListener('submit', (event) => {
    event.preventDefault();
    void load();
  });

  void load();
}

function skeletons() {
  return Array.from({ length: 3 }, () => `
    <div class="skeleton-card">
      <div class="skeleton-line short"></div>
      <div class="skeleton-line"></div>
      <div class="skeleton-line"></div>
      <div class="skeleton-line medium"></div>
    </div>
  `).join('');
}
