import { getAnalyticsResults, getMortgageQuote } from './api';
import { escapeHtml } from './main';

export function renderAnalytics() {
  const query = new URLSearchParams(location.hash.split('?')[1] || '');
  const address = query.get('address') || 'Москва, Тверская 1';
  return `
    <section class="page-heading">
      <span class="eyebrow">Аналитика</span>
      <h1>Анализ объекта</h1>
      <p>Сопоставляем адрес, инфраструктуру, курсы и ипотеку на реальных данных.</p>
    </section>
    <section class="tool-panel">
      <form id="analytics-form" class="parser-form">
        <div class="parser-row analytics-row">
          <label class="wide">
            <span>Адрес</span>
            <input id="analytics-address" name="address" value="${escapeHtml(address)}" placeholder="Москва, Тверская 1" />
          </label>
          <button class="button primary" type="submit">Проанализировать</button>
        </div>
      </form>
      <div id="analytics-results" class="analytics-results"></div>
    </section>
  `;
}

export function bindAnalytics() {
  const form = document.querySelector<HTMLFormElement>('#analytics-form');
  const results = document.querySelector<HTMLElement>('#analytics-results');
  if (!form || !results) return;

  const load = async () => {
    const address = new FormData(form).get('address')?.toString().trim() || '';
    if (!address) {
      results.innerHTML = '<div class="empty">Укажите адрес для анализа.</div>';
      return;
    }
    results.innerHTML = skeletons();
    try {
      const analytics = await getAnalyticsResults(address);
      const mortgage = await getMortgageQuote(15000000, 20, 20);
      const lat = Number(analytics.coordinates?.lat || 0);
      const lon = Number(analytics.coordinates?.lon || 0);
      const bbox = `${(lon - 0.008).toFixed(6)},${(lat - 0.004).toFixed(6)},${(lon + 0.008).toFixed(6)},${(lat + 0.004).toFixed(6)}`;
      const osmMap = `https://www.openstreetmap.org/export/embed.html?bbox=${bbox}&layer=mapnik&marker=${lat.toFixed(6)},${lon.toFixed(6)}`;
      const osmLink = `https://www.openstreetmap.org/?mlat=${lat.toFixed(6)}&mlon=${lon.toFixed(6)}#map=16/${lat.toFixed(6)}/${lon.toFixed(6)}`;
      const buildingsText = (analytics.buildings || []).length ? (analytics.buildings || []).slice(0, 3).map((item) => `<li>${escapeHtml(String(item.name || 'Здание'))}${item.address ? ` — ${escapeHtml(String(item.address))}` : ''}</li>`).join('') : '<li>Кадастровый номер и стоимость — в выписке ЕГРН или на публичной кадастровой карте.</li>';
      results.innerHTML = `
        <div class="analytics-layout">
          <div class="analytics-main">
            <div class="summary-grid">
              <div class="metric-card">
                <span>Нормализованный адрес</span>
                <strong>${escapeHtml(analytics.normalized || analytics.address)}</strong>
              </div>
              <div class="metric-card">
                <span>Координаты</span>
                <strong>${lat.toFixed(4)}, ${lon.toFixed(4)}</strong>
              </div>
              <div class="metric-card">
                <span>Источники</span>
                <strong>${(analytics.sources_used || []).join(', ') || 'nominatim'}</strong>
              </div>
            </div>
            <div class="infra-grid">
              ${renderInfraCard('🚇', 'Метро', analytics.infrastructure?.metro || [])}
              ${renderInfraCard('🏫', 'Школы', analytics.infrastructure?.schools || [])}
              ${renderInfraCard('🌳', 'Парки', analytics.infrastructure?.parks || [])}
              ${renderInfraCard('💊', 'Аптеки', analytics.infrastructure?.pharmacies || [])}
            </div>
          </div>
          <div class="analytics-side">
            <div class="map-card">
              <iframe title="Объект на карте" src="${osmMap}" width="100%" height="300" style="border:0;border-radius:12px;" loading="lazy" referrerpolicy="no-referrer-when-downgrade"></iframe>
              <div class="map-meta">
                <span>© OpenStreetMap contributors</span>
                <a href="${osmLink}" target="_blank" rel="noreferrer">Открыть в OSM</a>
              </div>
            </div>
            <div class="finance-card">
              <div class="finance-strip">
                <span>USD ${Number(analytics.currency?.usd || 0).toFixed(2)} ₽</span>
                <span>EUR ${Number(analytics.currency?.eur || 0).toFixed(2)} ₽</span>
                <span>Ставка ${Number(analytics.mortgage?.key_rate || 16.5).toFixed(1)}%</span>
              </div>
              <div class="compact-mortgage">
                <input id="mortgage-price" type="number" value="15000000" />
                <input id="mortgage-years" type="number" value="20" />
                <input id="mortgage-down" type="number" value="20" />
                <button type="button" class="button primary" id="mortgage-calc">Рассчитать</button>
                <div id="mortgage-result" class="mortgage-result">Ежемесячно: ${Number(mortgage.monthly || 0).toLocaleString('ru-RU', {maximumFractionDigits: 2})} ₽</div>
              </div>
              <div class="cadastre-card">
                <h3>Кадастровая информация</h3>
                <ul>${buildingsText}</ul>
                <a class="button alt" href="https://pkk.rosreestr.ru/" target="_blank" rel="noreferrer">Публичная кадастровая карта</a>
              </div>
            </div>
          </div>
        </div>
      `;
      const calcBtn = document.querySelector<HTMLButtonElement>('#mortgage-calc');
      calcBtn?.addEventListener('click', async () => {
        const priceInput = document.querySelector<HTMLInputElement>('#mortgage-price');
        const yearsInput = document.querySelector<HTMLInputElement>('#mortgage-years');
        const downInput = document.querySelector<HTMLInputElement>('#mortgage-down');
        if (!priceInput || !yearsInput || !downInput) return;
        const mortgageData = await getMortgageQuote(Number(priceInput.value || 0), Number(yearsInput.value || 0), Number(downInput.value || 0));
        const result = document.querySelector<HTMLElement>('#mortgage-result');
        if (result) {
          result.innerHTML = `Ежемесячно: ${Number(mortgageData.monthly || 0).toLocaleString('ru-RU', { maximumFractionDigits: 2 })} ₽<br>Всего: ${Number(mortgageData.total || 0).toLocaleString('ru-RU', { maximumFractionDigits: 2 })} ₽`;
        }
      });
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

function renderInfraCard(icon: string, label: string, items: Array<{ name: string; distance_m: number; address?: string }>) {
  return `
    <div class="metric-card infra-card">
      <span>${icon} ${label}</span>
      <strong>${items.length || 0}</strong>
      <div class="infra-list">
        ${items.length ? items.slice(0, 3).map((item) => `<div>${escapeHtml(item.name)} · ${Math.round(item.distance_m)} м</div>`).join('') : '<div>Не найдено рядом</div>'}
      </div>
    </div>
  `;
}

function skeletons() {
  return '<div class="skeleton-card"><div class="skeleton-line short"></div><div class="skeleton-line"></div><div class="skeleton-line"></div></div>';
}
