<script setup>
import { ref, onMounted } from 'vue';
import { loginWithEmail, registerWithEmail, getUserData } from '../firebase';

const emit = defineEmits(['logged-in', 'waiting-approval']);

const email = ref('');
const password = ref('');
const rememberMe = ref(true);
const isLoading = ref(false);
const errorMsg = ref('');
const isLoginMode = ref(true);

const isValidEmail = (e) => /^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(e);

onMounted(() => {
  const savedEmail = localStorage.getItem('jdsa_email');
  const savedPassword = localStorage.getItem('jdsa_password');
  if (savedEmail) email.value = savedEmail;
  if (savedPassword) password.value = savedPassword;
});

const handleSubmit = async () => {
  errorMsg.value = '';
  if (!email.value.trim() || !password.value.trim()) {
    errorMsg.value = 'Por favor, completá todos los campos.';
    return;
  }
  if (!isValidEmail(email.value)) {
    errorMsg.value = 'El email no es válido.';
    return;
  }
  if (password.value.length < 6) {
    errorMsg.value = 'La contraseña debe tener al menos 6 caracteres.';
    return;
  }

  isLoading.value = true;
  try {
    let user;
    if (isLoginMode.value) {
      user = await loginWithEmail(email.value, password.value);
    } else {
      user = await registerWithEmail(email.value, password.value);
    }
    
    const userData = await getUserData(user.uid);

    if (userData?.is_enabled) {
      if (isLoginMode.value && rememberMe.value) {
        localStorage.setItem('jdsa_email', email.value);
        localStorage.setItem('jdsa_password', password.value);
      } else if (!rememberMe.value) {
        localStorage.removeItem('jdsa_email');
        localStorage.removeItem('jdsa_password');
      }
      emit('logged-in', { user, userData });
    } else {
      emit('waiting-approval', { user, userData });
    }
  } catch (err) {
    if (err.code === 'auth/invalid-credential' || err.code === 'auth/wrong-password' || err.code === 'auth/user-not-found') {
      errorMsg.value = 'Email o contraseña incorrectos.';
    } else if (err.code === 'auth/email-already-in-use') {
      errorMsg.value = 'El email ya está registrado.';
    } else {
      errorMsg.value = 'Error: ' + err.message;
    }
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
        {{ isLoginMode ? 'Ingresá tus credenciales para acceder.' : 'Ingresá email y contraseña para solicitar acceso. Un administrador habilitará tu cuenta.' }}
      </p>

      <!-- Form inputs -->
      <div class="w-full flex flex-col gap-3">
        <input
          v-model="email"
          type="email"
          placeholder="tu@email.com"
          :disabled="isLoading"
          class="w-full px-4 py-3 rounded-xl border border-slate-200 dark:border-slate-600 bg-white dark:bg-slate-700 text-slate-800 dark:text-slate-200 focus:outline-none focus:ring-2 focus:ring-indigo-500 focus:border-transparent transition-all placeholder:text-slate-400 dark:placeholder:text-slate-500 disabled:opacity-50"
        />

        <input
          v-model="password"
          type="password"
          placeholder="Contraseña"
          :disabled="isLoading"
          @keyup.enter="handleSubmit"
          class="w-full px-4 py-3 rounded-xl border border-slate-200 dark:border-slate-600 bg-white dark:bg-slate-700 text-slate-800 dark:text-slate-200 focus:outline-none focus:ring-2 focus:ring-indigo-500 focus:border-transparent transition-all placeholder:text-slate-400 dark:placeholder:text-slate-500 disabled:opacity-50"
        />

        <div v-if="isLoginMode" class="flex items-center gap-2 mt-1 px-1">
          <input 
            type="checkbox" 
            id="rememberMe" 
            v-model="rememberMe"
            class="w-4 h-4 text-indigo-600 bg-slate-100 border-slate-300 rounded focus:ring-indigo-500 dark:focus:ring-indigo-600 dark:ring-offset-slate-800 focus:ring-2 dark:bg-slate-700 dark:border-slate-600"
          >
          <label for="rememberMe" class="text-sm font-medium text-slate-700 dark:text-slate-300 cursor-pointer select-none">Recordar mis datos localmente</label>
        </div>

        <button
          @click="handleSubmit"
          :disabled="isLoading || !email.trim() || !password.trim()"
          class="w-full flex items-center justify-center gap-3 px-6 py-4 bg-indigo-600 hover:bg-indigo-700 text-white rounded-xl transition-all shadow-md shadow-indigo-200 dark:shadow-indigo-900/30 active:scale-95 disabled:opacity-50 disabled:active:scale-100 font-semibold"
        >
          <span v-if="isLoading" class="animate-spin h-5 w-5 border-2 border-white/30 border-t-white rounded-full"></span>
          <svg v-else-if="!isLoginMode" xmlns="http://www.w3.org/2000/svg" class="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M16 12a4 4 0 10-8 0 4 4 0 008 0zm0 0v1.5a2.5 2.5 0 005 0V12a9 9 0 10-9 9m4.5-1.206a8.959 8.959 0 01-4.5 1.207" />
          </svg>
          <svg v-else xmlns="http://www.w3.org/2000/svg" class="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M11 16l-4-4m0 0l4-4m-4 4h14m-5 4v1a3 3 0 01-3 3H6a3 3 0 01-3-3V7a3 3 0 013-3h7a3 3 0 013 3v1" />
          </svg>
          {{ isLoading ? 'Procesando...' : (isLoginMode ? 'Iniciar Sesión' : 'Solicitar acceso') }}
        </button>

        <button 
          @click="isLoginMode = !isLoginMode" 
          :disabled="isLoading"
          class="mt-2 text-sm text-indigo-600 dark:text-indigo-400 hover:text-indigo-800 dark:hover:text-indigo-300 transition-colors bg-transparent border-none cursor-pointer"
        >
          {{ isLoginMode ? '¿No tenés cuenta? Solicitar acceso' : '¿Ya tenés cuenta? Iniciar Sesión' }}
        </button>
      </div>

      <p v-if="errorMsg" class="mt-4 text-sm text-red-500 dark:text-red-400 font-medium text-center bg-red-50 dark:bg-red-900/20 px-4 py-2 rounded-lg border border-red-100 dark:border-red-800 w-full animate-in fade-in slide-in-from-top-2">
        {{ errorMsg }}
      </p>

      <p class="mt-8 text-xs text-slate-400 dark:text-slate-500 uppercase tracking-widest font-bold">iCTG Assistant</p>
    </div>
  </div>
</template>
