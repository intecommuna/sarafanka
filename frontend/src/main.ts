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

export type SessionUser = { id: number; email: string; name: string; role: string; created_at: string };
export let currentUser: SessionUser | null = null;
export const setCurrentUser = (user: SessionUser | null) => { currentUser = user; };
const initialTheme = localStorage.getItem('theme') || 'light';
document.documentElement.setAttribute('data-theme', initialTheme);
const applyTheme = (theme: string) => { document.documentElement.setAttribute('data-theme', theme); localStorage.setItem('theme', theme); const toggle = document.querySelector<HTMLButtonElement>('#theme-toggle'); if (toggle) toggle.textContent = theme === 'dark' ? '☀️' : '🌙'; };

const app = document.querySelector<HTMLElement>('#app')!;
const render = (page: () => Promise<string> | string) => {
  app.innerHTML = '<div class="loading">Загрузка...</div>';
  Promise.resolve(page()).then((html) => { app.innerHTML = html; const hash = location.hash; if (hash.includes('/catalog')) bindCatalog(); if (hash.match(/^#\/ads\/\d+$/)) bindAd(Number(hash.split('/')[2])); if (hash === '#/ad/new') bindAdForm(); if (hash.includes('/ad/') && hash.includes('/edit')) bindAdForm(Number(hash.split('/')[2])); if (hash === '#/my') bindMy(); if (hash.match(/^#\/news\/\d+$/)) bindNewsItem(Number(hash.split('/')[2])); if (hash.includes('/news/new')) bindNewsForm(); if (hash.includes('/news/') && hash.includes('/edit')) bindNewsForm(Number(hash.split('/')[2])); if (hash === '#/login') bindAuth('login'); if (hash === '#/register') bindAuth('register'); });
};

export const refreshHeader = () => {
  const header = document.querySelector<HTMLElement>('#site-header');
  if (!header) return;
  header.innerHTML = `<div class="header-inner container">
    <a class="brand" href="#/"><span class="brand-mark">С</span><span>Сарафанка</span></a>
    <form class="search" id="search-form"><input id="header-search" placeholder="Найти объявление" aria-label="Поиск"><button aria-label="Искать">⌕</button></form>
    <button class="menu-toggle" id="menu-toggle" type="button" aria-label="Открыть меню" aria-expanded="false"><span></span><span></span><span></span></button>
    <nav id="main-nav"><a href="#/catalog">Каталог</a><a href="#/news">Новости</a>${currentUser ? `<a href="#/my">Мои объявления</a>${currentUser.role === 'admin' ? '<a href="#/admin">Админка</a>' : ''}${['admin','moderator'].includes(currentUser.role) ? '<a class="nav-add" href="#/news/new">+ Новость</a>' : ''}<span class="user-name">${escapeHtml(currentUser.name)}</span><button class="link-button" id="logout">Выйти</button>` : '<a href="#/login">Войти</a><a class="nav-add" href="#/register">Регистрация</a>'}<button class="theme-toggle" id="theme-toggle" type="button" aria-label="Переключить тему">${initialTheme === 'dark' ? '☀️' : '🌙'}</button></nav>
    <a class="button primary compact" href="${currentUser ? '#/ad/new' : '#/login'}">+ Объявление</a>
  </div>`;
  header.querySelector<HTMLFormElement>('#search-form')?.addEventListener('submit', (event) => { event.preventDefault(); location.hash = `#/catalog?q=${encodeURIComponent(header.querySelector<HTMLInputElement>('#header-search')!.value)}`; });
  header.querySelector('#logout')?.addEventListener('click', () => { localStorage.removeItem('token'); currentUser = null; refreshHeader(); location.hash = '#/'; });
  header.querySelector('#theme-toggle')?.addEventListener('click', () => applyTheme(document.documentElement.dataset.theme === 'dark' ? 'light' : 'dark'));
  const menuToggle = header.querySelector<HTMLButtonElement>('#menu-toggle');
  const nav = header.querySelector<HTMLElement>('#main-nav');
  menuToggle?.addEventListener('click', () => { const isOpen = nav?.classList.toggle('is-open') ?? false; menuToggle.setAttribute('aria-expanded', String(isOpen)); });
  nav?.querySelectorAll('a').forEach((link) => link.addEventListener('click', () => { nav.classList.remove('is-open'); menuToggle?.setAttribute('aria-expanded', 'false'); }));
};

export const escapeHtml = (value: string) => value.replace(/[&<>'"]/g, (char) => ({ '&': '&amp;', '<': '&lt;', '>': '&gt;', "'": '&#39;', '"': '&quot;' }[char]!));
export const requireUser = () => { if (!currentUser) { location.hash = '#/login'; return false; } return true; };
export const canEdit = (ownerId: number) => Boolean(currentUser && (currentUser.id === ownerId || ['moderator', 'admin'].includes(currentUser.role)));

refreshHeader();
initRouter({ '/': () => render(renderHome), '/catalog': (params) => render(() => renderCatalog(params)), '/ads/:id': ({ id }) => render(() => renderAd(Number(id))), '/ad/new': () => render(renderAdForm), '/ad/:id/edit': ({ id }) => render(() => renderAdForm(Number(id))), '/news': () => render(renderNews), '/news/new': () => render(renderNewsForm), '/news/:id': ({ id }) => render(() => renderNewsItem(Number(id))), '/news/:id/edit': ({ id }) => render(() => renderNewsForm(Number(id))), '/login': () => render(renderLogin), '/register': () => render(renderRegister), '/my': () => render(renderMy), '/admin': () => render(renderAdmin) });
if (localStorage.getItem('token')) me().then((user) => { currentUser = user; refreshHeader(); }).catch(() => { localStorage.removeItem('token'); });
