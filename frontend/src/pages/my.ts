import { deleteAd, getAds } from '../api';
import { card } from './home';
import { currentUser, escapeHtml, requireUser } from '../main';
export async function renderMy() { if (!requireUser()) return ''; try { const ads = (await getAds()).filter((ad) => ad.UserID === currentUser!.id); return `<section class="page-heading inline-heading"><div><span class="eyebrow">Личный кабинет</span><h1>Мои объявления</h1></div><a class="button primary" href="#/ad/new">+ Создать</a></section><div class="ad-grid">${ads.length ? ads.map(card).join('') : '<div class="empty">У вас пока нет объявлений</div>'}</div>`; } catch (error) { return `<div class="error">${escapeHtml((error as Error).message)}</div>`; } }
export function bindMy() { document.querySelectorAll<HTMLElement>('[data-delete-ad]').forEach((button) => button.addEventListener('click', async () => { await deleteAd(Number(button.dataset.deleteAd)); location.hash = '#/my'; })); }
