import './style.css';
import { me } from './api';
import { initRouter } from './router';
import { renderAdmin } from './pages/admin';
import { bindAd, bindAdForm, renderAd, renderAdForm } from './pages/ad';
import { bindCatalog, renderCatalog } from './pages/catalog';
import { renderHome } from './pages/home';
import { bindAuth, renderLogin, renderRegister } from './pages/auth';
import { bindMy, renderMy } from './pages/my';
import { bindNewsForm, bindNewsItem, renderNews, renderNewsForm, renderNewsItem } from './pages/news';
import { bindAnalytics, renderAnalytics } from './analytics';
import { bindTools, renderTools } from './tools';

export type SessionUser = { id: number; email: string; name: string; role: string; created_at: string };
export let currentUser: SessionUser | null = null;
export const setCurrentUser = (user: SessionUser | null) => { currentUser = user; };
const initialTheme = localStorage.getItem('theme') || 'light';
document.documentElement.setAttribute('data-theme', initialTheme);
const applyTheme = (theme: string) => { document.documentElement.setAttribute('data-theme', theme); localStorage.setItem('theme', theme); const toggle = document.querySelector<HTMLButtonElement>('#theme-toggle'); if (toggle) toggle.textContent = theme === 'dark' ? '☀️' : '🌙'; };

const app = document.querySelector<HTMLElement>('#app')!;
const render = (page: () => Promise<string> | string) => {
  app.innerHTML = '<div class="loading">Загрузка...</div>';
  Promise.resolve(page()).then((html) => { app.innerHTML = html; const hash = location.hash; if (hash.includes('/catalog')) bindCatalog(); if (hash.match(/^#\/ads\/\d+$/)) bindAd(Number(hash.split('/')[2])); if (hash === '#/ad/new') bindAdForm(); if (hash.includes('/ad/') && hash.includes('/edit')) bindAdForm(Number(hash.split('/')[2])); if (hash === '#/my') bindMy(); if (hash.match(/^#\/news\/\d+$/)) bindNewsItem(Number(hash.split('/')[2])); if (hash.includes('/news/new')) bindNewsForm(); if (hash.includes('/news/') && hash.includes('/edit')) bindNewsForm(Number(hash.split('/')[2])); if (hash === '#/login') bindAuth('login'); if (hash === '#/register') bindAuth('register'); if (hash === '#/tools') bindTools(); if (hash.startsWith('#/tools/analytics')) bindAnalytics(); });
};

type ExtendedWindow = Window & {
  __sarafnkaMenuOutsideClick?: (event: MouseEvent) => void;
  __sarafnkaResizeHandler?: () => void;
};

const refreshNav = () => {
  const hdr = document.querySelector<HTMLElement>('.hdr');
  const inEl = document.querySelector<HTMLElement>('.hdr-in');

  if (!hdr || !inEl) return;

  hdr.classList.remove('compact');
  requestAnimationFrame(() => {
    if (inEl.scrollWidth > inEl.clientWidth + 1) hdr.classList.add('compact');
  });
};

export const refreshHeader = () => {
  const header = document.querySelector<HTMLElement>('#site-header');
  if (!header) return;

  header.className = 'hdr';

  const isLoggedIn = Boolean(currentUser);
  const navLinks = currentUser
    ? `<a href="#/catalog">Каталог</a><a href="#/news">Новости</a><a href="#/tools">Инструменты</a><a href="#/my">Мои объявления</a>${currentUser.role === 'admin' ? '<a href="#/admin">Админка</a>' : ''}`
    : '<a href="#/catalog">Каталог</a><a href="#/news">Новости</a><a href="#/tools">Инструменты</a>';

  const logoutIcon = `
    <svg viewBox="0 0 24 24" aria-hidden="true" focusable="false">
      <path d="M9 21H5a2 2 0 0 1-2-2V5a2 2 0 0 1 2-2h4" />
      <path d="M16 17l5-5-5-5" />
      <path d="M21 12H9" />
    </svg>
  `;

  const mobileLinks = `
    <a href="#/catalog">Каталог</a>
    <a href="#/news">Новости</a>
    <a href="#/tools">Инструменты</a>
    ${currentUser ? '<a href="#/my">Мои объявления</a>' : ''}
    ${currentUser && currentUser.role === 'admin' ? '<a href="#/admin">Админка</a>' : ''}
    ${currentUser ? '<a href="#/ad/new">+ Объявление</a>' : '<a href="#/login">Войти</a><a href="#/register">Регистрация</a>'}
    ${currentUser ? `<span class="mnav-user">${escapeHtml(currentUser.name)}</span>` : ''}
    ${currentUser ? '<button class="mnav-logout" type="button" id="mnav-logout">Выйти</button>' : ''}
  `;

  header.innerHTML = `
    <div class="hdr-in">
      <a class="logo" href="#/">Сарафанка</a>
      <nav class="nav" id="main-nav">${navLinks}</nav>
      <div class="acts">
        <span class="only-wide uname">${isLoggedIn ? escapeHtml(currentUser!.name) : ''}</span>
        ${isLoggedIn ? '<a class="btn btn-sm only-wide" href="#/ad/new">+ Объявление</a>' : ''}
        ${isLoggedIn ? `<button class="ibtn only-wide" id="logout" type="button" aria-label="Выйти" title="Выйти">${logoutIcon}</button>` : ''}
        <button class="ibtn" id="theme-toggle" type="button" aria-label="Переключить тему">${document.documentElement.getAttribute('data-theme') === 'dark' ? '☀️' : '🌙'}</button>
        <button class="ibtn" id="search-toggle" type="button" aria-label="Открыть поиск">🔍</button>
        <button class="ibtn burger" id="menu-toggle" type="button" aria-label="Открыть меню">☰</button>
      </div>
    </div>
    <div class="search-panel" id="search-panel">
      <input id="search-input" type="search" placeholder="Найти объявление" />
      <button class="btn btn-sm" id="search-go" type="button">Найти</button>
    </div>
    <nav class="mnav" id="mnav">${mobileLinks}</nav>
  `;

  const themeToggle = header.querySelector<HTMLButtonElement>('#theme-toggle');
  const searchToggle = header.querySelector<HTMLButtonElement>('#search-toggle');
  const searchPanel = header.querySelector<HTMLElement>('#search-panel');
  const searchInput = header.querySelector<HTMLInputElement>('#search-input');
  const searchGo = header.querySelector<HTMLButtonElement>('#search-go');
  const mnav = header.querySelector<HTMLElement>('#mnav');
  const burger = header.querySelector<HTMLButtonElement>('#menu-toggle');
  const nav = header.querySelector<HTMLElement>('#main-nav');

  const closeMenu = () => {
    mnav?.classList.remove('open');
    burger?.setAttribute('aria-expanded', 'false');
  };

  const closeSearch = () => {
    searchPanel?.classList.remove('open');
  };

  const openSearch = () => {
    searchPanel?.classList.add('open');
    requestAnimationFrame(() => searchInput?.focus());
  };

  const handleDocumentClick = (event: MouseEvent) => {
    const target = event.target as Node;
    if (mnav && burger && !mnav.contains(target) && !burger.contains(target)) closeMenu();
    if (searchPanel && searchToggle && !searchPanel.contains(target) && !searchToggle.contains(target)) closeSearch();
  };

  const handleKeyDown = (event: KeyboardEvent) => {
    if (event.key === 'Escape') {
      closeMenu();
      closeSearch();
    }
  };

  if (!(header as HTMLElement & { __headerBound?: boolean }).__headerBound) {
    document.addEventListener('click', handleDocumentClick);
    document.addEventListener('keydown', handleKeyDown);
    (header as HTMLElement & { __headerBound?: boolean }).__headerBound = true;
  }

  themeToggle?.addEventListener('click', () => {
    const nextTheme = document.documentElement.getAttribute('data-theme') === 'dark' ? 'light' : 'dark';
    applyTheme(nextTheme);
    refreshHeader();
  });

  searchToggle?.addEventListener('click', (event) => {
    event.stopPropagation();
    if (searchPanel?.classList.contains('open')) {
      closeSearch();
      return;
    }
    openSearch();
  });

  searchGo?.addEventListener('click', () => {
    const value = searchInput?.value.trim() ?? '';
    if (!value) {
      closeSearch();
      return;
    }
    location.hash = `#/catalog?q=${encodeURIComponent(value)}`;
    closeSearch();
  });

  searchInput?.addEventListener('keydown', (event) => {
    if (event.key !== 'Enter') return;
    const value = searchInput.value.trim();
    if (!value) {
      closeSearch();
      return;
    }
    location.hash = `#/catalog?q=${encodeURIComponent(value)}`;
    closeSearch();
  });

  burger?.addEventListener('click', (event) => {
    event.stopPropagation();
    const isOpen = mnav?.classList.toggle('open');
    burger.setAttribute('aria-expanded', String(Boolean(isOpen)));
  });

  mnav?.querySelectorAll('a, button').forEach((link) => {
    link.addEventListener('click', () => {
      closeMenu();
    });
  });

  header.querySelector('#logout')?.addEventListener('click', () => {
    localStorage.removeItem('token');
    currentUser = null;
    refreshHeader();
    location.hash = '#/';
  });

  header.querySelector('#mnav-logout')?.addEventListener('click', () => {
    localStorage.removeItem('token');
    currentUser = null;
    refreshHeader();
    location.hash = '#/';
  });

  nav?.querySelectorAll('a').forEach((link) => {
    link.addEventListener('click', () => {
      closeMenu();
    });
  });

  const handleResize = () => refreshNav();
  const w = window as ExtendedWindow;
  window.removeEventListener('resize', w.__sarafnkaResizeHandler ?? (() => undefined));
  w.__sarafnkaResizeHandler = handleResize;
  window.addEventListener('resize', handleResize);

  refreshNav();
};

let resizeTimer: number | undefined;
const handleResizeDebounced = () => {
  if (resizeTimer) window.clearTimeout(resizeTimer);
  resizeTimer = window.setTimeout(() => refreshNav(), 100);
};

window.addEventListener('resize', handleResizeDebounced);

export const escapeHtml = (value: string) => value.replace(/[&<>'"]/g, (char) => ({ '&': '&amp;', '<': '&lt;', '>': '&gt;', "'": '&#39;', '"': '&quot;' }[char]!));
export const requireUser = () => { if (!currentUser) { location.hash = '#/login'; return false; } return true; };
export const canEdit = (ownerId: number) => Boolean(currentUser && (currentUser.id === ownerId || ['moderator', 'admin'].includes(currentUser.role)));

refreshHeader();
initRouter({
  '/': () => render(renderHome),
  '/catalog': (params) => render(() => renderCatalog(params)),
  '/catalog/category/:slug': ({ slug }) => render(() => renderCatalog({ type: 'product', category: slug })),
  '/catalog/services': () => render(() => renderCatalog({ type: 'service' })),
  '/catalog/services/:slug': ({ slug }) => render(() => renderCatalog({ type: 'service', category: slug })),
  '/ads/:id': ({ id }) => render(() => renderAd(Number(id))),
  '/ad/new': () => render(renderAdForm),
  '/ad/:id/edit': ({ id }) => render(() => renderAdForm(Number(id))),
  '/news': () => render(renderNews),
  '/news/new': () => render(renderNewsForm),
  '/news/:id': ({ id }) => render(() => renderNewsItem(Number(id))),
  '/news/:id/edit': ({ id }) => render(() => renderNewsForm(Number(id))),
  '/login': () => render(renderLogin),
  '/register': () => render(renderRegister),
  '/my': () => render(renderMy),
  '/admin': () => render(renderAdmin),
  '/tools': () => render(() => {
    if (!requireUser()) return '';
    return renderTools();
  }),
  '/tools/parser': () => render(() => {
    if (!requireUser()) return '';
    return renderTools();
  }),
  '/tools/analytics': () => render(() => {
    if (!requireUser()) return '';
    return renderAnalytics();
  })
});
if (localStorage.getItem('token')) me().then((user) => { currentUser = user; refreshHeader(); }).catch(() => { localStorage.removeItem('token'); });  const toolRoute = location.hash.startsWith('#/tools');
  if (toolRoute && !currentUser) {
    location.hash = '#/login';
  }