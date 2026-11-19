<script>
  import { onMount } from 'svelte';
  import ArrowLeft from 'lucide-svelte/icons/arrow-left';
  import Boxes from 'lucide-svelte/icons/boxes';
  import ChartColumn from 'lucide-svelte/icons/chart-column';
  import Clock from 'lucide-svelte/icons/clock';
  import LayoutDashboard from 'lucide-svelte/icons/layout-dashboard';
  import LogOut from 'lucide-svelte/icons/log-out';
  import Package from 'lucide-svelte/icons/package';
  import Receipt from 'lucide-svelte/icons/receipt';
  import Settings from 'lucide-svelte/icons/settings';
  import Sparkles from 'lucide-svelte/icons/sparkles';
  import Target from 'lucide-svelte/icons/target';
  import Truck from 'lucide-svelte/icons/truck';
  import Wallet from 'lucide-svelte/icons/wallet';
  import StylePicker from '../components/StylePicker.svelte';
  import Wordmark from '../components/Wordmark.svelte';
  import adminStrings from './strings.js';
  import Login from './Login.svelte';
  import Overview from './Overview.svelte';
  import Products from './Products.svelte';
  import Restock from './Restock.svelte';
  import Stock from './Stock.svelte';
  import Shifts from './Shifts.svelte';
  import Cash from './Cash.svelte';
  import Sales from './Sales.svelte';
  import RentalsAdmin from './RentalsAdmin.svelte';
  import Reports from './Reports.svelte';
  import Recap from './Recap.svelte';
  import SettingsPage from './SettingsPage.svelte';
  import { get, post, whenSessionLost } from '../lib/api.js';
  import { addStrings, getLanguage, t } from '../lib/i18n.svelte.js';
  import { route } from '../lib/route.svelte.js';
  import { pickLanguage, refresh, till, toastError } from '../lib/state.svelte.js';

  addStrings(adminStrings);

  const pages = [
    { path: '/admin', label: 'admin.overview', icon: LayoutDashboard, page: Overview },
    { path: '/admin/products', label: 'admin.products', icon: Package, page: Products },
    { path: '/admin/restock', label: 'admin.restock', icon: Truck, page: Restock },
    { path: '/admin/stock', label: 'admin.stock', icon: Boxes, page: Stock },
    { path: '/admin/shifts', label: 'admin.shifts', icon: Clock, page: Shifts },
    { path: '/admin/cash', label: 'admin.cash', icon: Wallet, page: Cash },
    { path: '/admin/sales', label: 'admin.sales', icon: Receipt, page: Sales },
    { path: '/admin/rentals', label: 'admin.rentals', icon: Target, page: RentalsAdmin },
    { path: '/admin/reports', label: 'admin.reports', icon: ChartColumn, page: Reports },
    { path: '/admin/recap', label: 'admin.recap', icon: Sparkles, page: Recap },
    { path: '/admin/settings', label: 'admin.settings', icon: Settings, page: SettingsPage },
  ];

  let session = $state(null);
  const current = $derived(pages.find((p) => p.path === route.path) ?? pages[0]);

  async function load() {
    try {
      session = await get('/api/admin/session');
      if (session.loggedIn && !till.loaded) await refresh();
    } catch (err) {
      toastError(err);
    }
  }

  async function logout() {
    await post('/api/admin/logout').catch(toastError);
    await load();
  }

  async function backToTill() {
    route.go('/');
    await refresh().catch(toastError);
  }

  onMount(() => {
    whenSessionLost(() => {
      if (session) session.loggedIn = false;
    });
    load();
  });
</script>

{#if !session}
  <div class="loading muted">{t('common.loading')}</div>
{:else if !session.loggedIn}
  <Login {session} onlogin={load} onback={session.local ? backToTill : null} />
{:else}
  <div class="admin ledger">
    <nav class="side machine" aria-label={t('login.title')}>
      <div class="brand">
        <Wordmark />
        <span class="muted small">{t('login.title')}</span>
      </div>
      {#each pages as item}
        <button class="link" class:active={current === item} onclick={() => route.go(item.path)} aria-current={current === item ? 'page' : undefined}>
          <item.icon size={20} />
          <span>{t(item.label)}</span>
        </button>
      {/each}
      <div class="grow"></div>
      <div class="tools">
        <button class="btn ghost" aria-label={t('set.language')} title={t('set.language')} onclick={() => pickLanguage(getLanguage() === 'de' ? 'en' : 'de')}>
          {getLanguage() === 'de' ? 'EN' : 'DE'}
        </button>
        <StylePicker beside />
      </div>
      {#if session.local}
        <button class="link" onclick={backToTill}><ArrowLeft size={20} /><span>{t('admin.backToTill')}</span></button>
      {/if}
      <button class="link" onclick={logout}><LogOut size={20} /><span>{t('admin.logout')}</span></button>
    </nav>
    <main class="page">
      {#key current}
        <current.page />
      {/key}
    </main>
  </div>
{/if}

<style>
  .loading { display: flex; align-items: center; justify-content: center; height: 100%; }
  .admin { display: grid; grid-template-columns: 216px minmax(0, 1fr); height: 100%; }
  .side {
    display: flex;
    flex-direction: column;
    gap: 2px;
    padding: 12px 0 12px 12px;
    background: var(--bg);
    overflow-y: auto;
  }
  .brand { display: flex; flex-direction: column; gap: 4px; padding: 8px 10px 16px; }
  .link {
    display: flex;
    align-items: center;
    gap: 12px;
    flex-shrink: 0;
    min-height: 44px;
    padding: 0 10px;
    border: 0;
    border-radius: 6px 0 0 6px;
    background: transparent;
    color: var(--muted);
    font-weight: 600;
    text-align: left;
  }
  .link:hover { background: var(--surface-2); color: var(--text); }
  .tools { display: flex; flex-shrink: 0; align-items: center; gap: 6px; padding: 8px 10px; }
  .tools .btn { flex: none; width: 44px; margin-bottom: 0; padding: 0; color: var(--muted); }
  .tools .btn:hover { color: var(--text); }
  /* The open page is a tab of the book, joined to the paper beside it. */
  .link.active { background: var(--tab-bg); color: var(--tab-ink); }
  .page { min-width: 0; padding: 20px 24px; background: var(--bg); overflow-y: auto; }
  /* The double rule a cash book puts under a total. */
  .page :global(h1), .page :global(header.spread) { padding-bottom: 8px; border-bottom: 3px double var(--text); }
  .page :global(header.spread h1) { padding-bottom: 0; border-bottom: 0; }

  @media (max-width: 760px) {
    .admin { grid-template-columns: 56px minmax(0, 1fr); }
    .link span, .brand { display: none; }
    .tools { flex-direction: column; padding: 4px 0; }
    .side { padding: 6px 0 6px 6px; }
    .link { justify-content: center; padding: 0; }
    .page { padding: 16px; }
  }
</style>
