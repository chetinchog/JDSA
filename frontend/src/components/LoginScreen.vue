<script setup>
import { ref } from 'vue';
import { registerWithEmail, getUserData } from '../firebase';

const emit = defineEmits(['logged-in', 'waiting-approval']);

const email = ref('');
const isLoading = ref(false);
const errorMsg = ref('');

const isValidEmail = (e) => /^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(e);

const handleRegister = async () => {
  errorMsg.value = '';
  if (!email.value.trim()) {
    errorMsg.value = 'Por favor, ingresá tu email.';
    return;
  }
  if (!isValidEmail(email.value)) {
    errorMsg.value = 'El email no es válido.';
    return;
  }

  isLoading.value = true;
  try {
    const user = await registerWithEmail(email.value);
    const userData = await getUserData(user.uid);

    if (userData?.is_enabled) {
      emit('logged-in', { user, userData });
    } else {
      emit('waiting-approval', { user, userData });
    }
  } catch (err) {
    errorMsg.value = 'Error al registrarse: ' + err.message;
  } finally {
    isLoading.value = false;
  }
};
</script>

<template>
  <div class="h-screen w-full flex flex-col items-center justify-center bg-slate-50 dark:bg-slate-900 transition-colors p-6">
    <div class="max-w-md w-full bg-white dark:bg-slate-800 rounded-3xl p-8 shadow-xl shadow-slate-200/50 dark:shadow-black/20 border border-slate-100 dark:border-slate-700 flex flex-col items-center animate-in zoom-in-95 duration-500">

      <!-- Logo -->
      <div class="h-20 w-20 bg-indigo-600 rounded-2xl flex items-center justify-center shadow-lg shadow-indigo-200 dark:shadow-indigo-900/40 mb-6 relative">
        <div class="absolute inset-0 bg-indigo-500/30 rounded-2xl animate-ping opacity-75"></div>
        <svg xmlns="http://www.w3.org/2000/svg" class="h-10 w-10 text-white relative z-10" fill="none" viewBox="0 0 24 24" stroke="currentColor">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z" />
        </svg>
      </div>

      <h1 class="text-3xl font-bold text-slate-800 dark:text-slate-100 mb-2 tracking-tight">Bienvenido a JDSA</h1>
      <p class="text-slate-500 dark:text-slate-400 text-center text-sm mb-8 leading-relaxed">
        Ingresá tu email para solicitar acceso. Un administrador habilitará tu cuenta.
      </p>

      <!-- Email input -->
      <div class="w-full flex flex-col gap-3">
        <input
          v-model="email"
          type="email"
          placeholder="tu@email.com"
          :disabled="isLoading"
          @keyup.enter="handleRegister"
          class="w-full px-4 py-3 rounded-xl border border-slate-200 dark:border-slate-600 bg-white dark:bg-slate-700 text-slate-800 dark:text-slate-200 focus:outline-none focus:ring-2 focus:ring-indigo-500 focus:border-transparent transition-all placeholder:text-slate-400 dark:placeholder:text-slate-500 disabled:opacity-50"
        />

        <button
          @click="handleRegister"
          :disabled="isLoading || !email.trim()"
          class="w-full flex items-center justify-center gap-3 px-6 py-4 bg-indigo-600 hover:bg-indigo-700 text-white rounded-xl transition-all shadow-md shadow-indigo-200 dark:shadow-indigo-900/30 active:scale-95 disabled:opacity-50 disabled:active:scale-100 font-semibold"
        >
          <span v-if="isLoading" class="animate-spin h-5 w-5 border-2 border-white/30 border-t-white rounded-full"></span>
          <svg v-else xmlns="http://www.w3.org/2000/svg" class="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M16 12a4 4 0 10-8 0 4 4 0 008 0zm0 0v1.5a2.5 2.5 0 005 0V12a9 9 0 10-9 9m4.5-1.206a8.959 8.959 0 01-4.5 1.207" />
          </svg>
          {{ isLoading ? 'Registrando...' : 'Solicitar acceso' }}
        </button>
      </div>

      <p v-if="errorMsg" class="mt-4 text-sm text-red-500 dark:text-red-400 font-medium text-center bg-red-50 dark:bg-red-900/20 px-4 py-2 rounded-lg border border-red-100 dark:border-red-800 w-full animate-in fade-in slide-in-from-top-2">
        {{ errorMsg }}
      </p>

      <p class="mt-8 text-xs text-slate-400 dark:text-slate-500 uppercase tracking-widest font-bold">iCTG Assistant</p>
    </div>
  </div>
</template>
