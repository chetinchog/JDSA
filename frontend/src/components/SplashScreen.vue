<script setup>
import { onMounted, ref } from 'vue'

const emit = defineEmits(['finish'])
const visible = ref(true)

onMounted(() => {
  // Total animation time (~2.5s for drawing + 0.5s pause)
  setTimeout(() => {
    visible.value = false
    setTimeout(() => {
      emit('finish')
    }, 800) // Buffer for the fade-out transition
  }, 3200)
})
</script>

<template>
  <Transition name="splash-fade">
    <div v-if="visible" class="splash-container">
      <div class="logo-wrapper">
        <svg viewBox="0 0 400 120" class="jdsa-logo">
          <defs>
            <linearGradient id="gradient" x1="0%" y1="0%" x2="100%" y2="0%">
              <stop offset="0%" style="stop-color:#6366f1;stop-opacity:1" /> <!-- Indigo-500 -->
              <stop offset="50%" style="stop-color:#a855f7;stop-opacity:1" /> <!-- Purple-500 -->
              <stop offset="100%" style="stop-color:#10b981;stop-opacity:1" /> <!-- Emerald-500 -->
            </linearGradient>
            
            <filter id="glow" x="-20%" y="-20%" width="140%" height="140%">
              <feGaussianBlur stdDeviation="3" result="blur" />
              <feComposite in="SourceGraphic" in2="blur" operator="over" />
            </filter>
          </defs>

          <!-- J -->
          <path class="letter" d="M30,20 h40 v60 q0,20 -20,20 t-20,-20 v-10" />
          
          <!-- D -->
          <path class="letter" d="M100,20 v80 h20 q40,0 40,-40 t-40,-40 h-20" />
          
          <!-- S -->
          <path class="letter" d="M250,20 h-30 q-20,0 -20,20 t20,20 h20 q20,0 20,20 t-20,20 h-30" />
          
          <!-- A -->
          <path class="letter" d="M300,100 L330,20 L360,100 M315,65 h30" />
        </svg>
      </div>
      
      <div class="loading-state">
        <div class="shimmer"></div>
        <p class="loading-text">Iniciando JDSA Assistant...</p>
      </div>

      <!-- Decor elements -->
      <div class="blob blob-1"></div>
      <div class="blob blob-2"></div>
    </div>
  </Transition>
</template>

<style scoped>
.splash-container {
  position: fixed;
  inset: 0;
  z-index: 9999;
  background-color: #0f172a; /* Always dark for premium feel */
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  overflow: hidden;
}

.logo-wrapper {
  position: relative;
  width: 100%;
  max-width: 500px;
  padding: 0 40px;
  filter: drop-shadow(0 0 20px rgba(99, 102, 241, 0.3));
}

.jdsa-logo {
  width: 100%;
  height: auto;
  fill: none;
  stroke: url(#gradient);
  stroke-width: 8;
  stroke-linecap: round;
  stroke-linejoin: round;
}

.letter {
  stroke-dasharray: 400;
  stroke-dashoffset: 400;
  animation: draw 2s cubic-bezier(0.65, 0, 0.35, 1) forwards;
}

/* Stagger letters */
.letter:nth-child(1) { animation-delay: 0s; }
.letter:nth-child(2) { animation-delay: 0.3s; }
.letter:nth-child(3) { animation-delay: 0.6s; }
.letter:nth-child(4) { animation-delay: 0.9s; }

@keyframes draw {
  to {
    stroke-dashoffset: 0;
  }
}

.loading-state {
  margin-top: 40px;
  position: relative;
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 12px;
}

.loading-text {
  color: #94a3b8;
  font-size: 11px;
  font-weight: 600;
  text-transform: uppercase;
  letter-spacing: 0.2em;
  opacity: 0;
  animation: fadeIn 0.8s ease-out 2.2s forwards;
}

@keyframes fadeIn {
  from { opacity: 0; transform: translateY(10px); }
  to { opacity: 1; transform: translateY(0); }
}

.shimmer {
  width: 120px;
  height: 2px;
  background: #1e293b;
  position: relative;
  overflow: hidden;
  border-radius: 4px;
}

.shimmer::after {
  content: "";
  position: absolute;
  inset: 0;
  background: linear-gradient(90deg, transparent, #6366f1, transparent);
  transform: translateX(-100%);
  animation: shimmer-anim 1.5s infinite;
}

@keyframes shimmer-anim {
  100% { transform: translateX(100%); }
}

/* Blobs */
.blob {
  position: absolute;
  border-radius: 50%;
  filter: blur(80px);
  z-index: -1;
  opacity: 0.3;
}

.blob-1 {
  width: 300px;
  height: 300px;
  background: #6366f1;
  top: -150px;
  right: -100px;
  animation: blob-float 10s infinite alternate;
}

.blob-2 {
  width: 250px;
  height: 250px;
  background: #10b981;
  bottom: -100px;
  left: -100px;
  animation: blob-float 8s infinite alternate-reverse;
}

@keyframes blob-float {
  from { transform: translate(0, 0); }
  to { transform: translate(40px, 40px); }
}

/* Page Transition */
.splash-fade-leave-active {
  transition: all 0.8s cubic-bezier(0.65, 0, 0.35, 1);
}

.splash-fade-leave-to {
  opacity: 0;
  transform: scale(1.1);
  filter: blur(10px);
}
</style>
