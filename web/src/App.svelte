<script lang="ts">
  import FinanceView from './lib/components/finance/FinanceView.svelte';

  type Module = 'hub' | 'finance' | 'system';
  let activeModule = $state<Module>('finance');
</script>

<div class="min-h-screen bg-slate-950 text-slate-100 antialiased">
  <!-- Top Navigation Header -->
  <header class="border-b border-slate-800 bg-slate-900/50 backdrop-blur-md">
    <div class="mx-auto flex max-w-7xl items-center justify-between px-6 py-3">
      <div class="flex items-center space-x-3">
        <span class="rounded bg-blue-600 px-2 py-0.5 text-xs font-black tracking-widest text-white">PLOS</span>
        <span class="text-xs text-slate-500">v1.0.0-ce</span>
      </div>

      <nav class="flex space-x-1">
        <button
          class="rounded-lg px-3 py-1.5 text-sm font-medium transition-colors {activeModule === 'hub'
            ? 'bg-slate-800 text-white'
            : 'text-slate-400 hover:text-slate-200'}"
          onclick={() => (activeModule = 'hub')}
        >
          Hub
        </button>
        <button
          class="rounded-lg px-3 py-1.5 text-sm font-medium transition-colors {activeModule === 'finance'
            ? 'bg-slate-800 text-white'
            : 'text-slate-400 hover:text-slate-200'}"
          onclick={() => (activeModule = 'finance')}
        >
          Finance
        </button>
      </nav>
    </div>
  </header>

  <!-- Main Content Container -->
  <main class="mx-auto max-w-7xl p-6">
    {#if activeModule === 'hub'}
      <div class="grid grid-cols-1 gap-6 md:grid-cols-2 lg:grid-cols-3">
        <button
          onclick={() => (activeModule = 'finance')}
          class="group rounded-xl border border-slate-800 bg-slate-900 p-6 text-left transition-all hover:border-blue-500/50 hover:bg-slate-800/50"
        >
          <div class="flex items-center justify-between">
            <h2 class="text-lg font-bold text-slate-100 group-hover:text-blue-400">Finance & Cashflow</h2>
            <span class="rounded-full bg-emerald-500/10 px-2 py-0.5 text-xs font-semibold text-emerald-400">Active</span>
          </div>
          <p class="mt-2 text-xs text-slate-400">
            Cascade debt payoff engine, multi-currency accounts, and cashflow projections.
          </p>
        </button>
      </div>
    {:else if activeModule === 'finance'}
      <FinanceView />
    {/if}
  </main>
</div>
