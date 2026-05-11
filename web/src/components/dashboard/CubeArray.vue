<template>
  <div class="cube-scene" role="img" aria-label="3D cube grid representing machine connectivity status">
    <div class="cube-viewport">
      <div class="cube-grid">
        <div
          v-for="(cell, i) in displayCubes"
          :key="i"
          class="cube-cell"
          :style="cellPositionStyle(i)"
          @mouseenter="hoveredIndex = i"
          @mouseleave="hoveredIndex = null"
        >
          <div
            class="cube"
            :class="['cube--' + cell.status, { 'cube--hovered': hoveredIndex === i }]"
            :style="{ animationDelay: cellAnimationDelay(i) + 's' }"
          >
            <div class="face face-top"></div>
            <div class="face face-front"></div>
            <div class="face face-right"></div>
          </div>

          <!-- Hover Tooltip -->
          <Transition name="fade">
            <div 
              v-if="hoveredIndex === i && cell.status !== 'empty'" 
              class="cube-tooltip"
            >
              <div class="tooltip-name">{{ cell.name || (cell.status === 'local' ? 'Local Host' : 'Unknown Machine') }}</div>
              <div class="tooltip-status" :class="cell.status">
                <span class="status-dot"></span>
                {{ cell.status.toUpperCase() }}
              </div>
            </div>
          </Transition>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, ref } from 'vue'

export interface CubeData {
  id?: string
  name?: string
  status: 'online' | 'offline' | 'empty' | 'local'
}

interface Props {
  cubes?: CubeData[]
  maxCubes?: number
}

const props = withDefaults(defineProps<Props>(), {
  cubes: () => [],
  maxCubes: 16
})

const hoveredIndex = ref<number | null>(null)

const displayCubes = computed<CubeData[]>(() => {
  const result: CubeData[] = []
  for (let i = 0; i < props.maxCubes; i++) {
    result.push(props.cubes[i] ?? { status: 'empty' })
  }
  return result
})

const GRID_COLS = 4
const OFFSET = 48

function cellPositionStyle(index: number): Record<string, string> {
  const row = Math.floor(index / GRID_COLS)
  const col = index % GRID_COLS
  const center = (GRID_COLS - 1) / 2
  const x = (col - center) * OFFSET
  const y = (row - center) * OFFSET
  return {
    '--cell-x': `${x}px`,
    '--cell-y': `${y}px`
  }
}

function cellAnimationDelay(index: number): number {
  const row = Math.floor(index / GRID_COLS)
  const col = index % GRID_COLS
  return (row * GRID_COLS + col) * 0.12
}
</script>

<style scoped>
.cube-scene {
  display: flex;
  justify-content: center;
  align-items: center;
  padding: var(--space-8);
  min-width: 320px;
  min-height: 360px;
}

.cube-viewport {
  perspective: 1200px;
  perspective-origin: 50% 35%;
  width: 320px;
  height: 320px;
  display: flex;
  justify-content: center;
  align-items: center;
}

.cube-grid {
  position: relative;
  width: 160px;
  height: 160px;
  transform-style: preserve-3d;
  transform: rotateX(54deg) rotateZ(-45deg);
}

.cube-cell {
  position: absolute;
  top: 50%;
  left: 50%;
  transform-style: preserve-3d;
  transform: translate(-50%, -50%) translate(var(--cell-x, 0px), var(--cell-y, 0px));
}

.cube {
  position: relative;
  width: 48px;
  height: 48px;
  transform-style: preserve-3d;
  animation: cube-float 5s ease-in-out infinite;
}

.face {
  position: absolute;
  width: 48px;
  height: 48px;
  backface-visibility: visible;
  border: 1px solid rgba(255, 255, 255, 0.05);
}

.face-top { transform: rotateX(90deg) translateZ(24px); }
.face-front { transform: translateZ(24px); }
.face-right { transform: rotateY(90deg) translateZ(24px); }

/* Online State - Dark Green Glow */
.cube--online .face {
  background: var(--cube-face-online);
  box-shadow: 0 0 15px var(--cube-glow-online);
}

.cube--online .face-top { opacity: 0.95; }
.cube--online .face-front { filter: brightness(0.8); }
.cube--online .face-right { filter: brightness(0.6); }

/* Offline State - Dark Red Glow */
.cube--offline .face {
  background: var(--cube-face-offline);
  box-shadow: 0 0 15px var(--cube-glow-offline);
}

.cube--offline .face-top { opacity: 0.95; }
.cube--offline .face-front { filter: brightness(0.8); }
.cube--offline .face-right { filter: brightness(0.6); }

/* Local State - Blue Glow (host machine) */
.cube--local .face {
  background: var(--cube-face-local);
  box-shadow: 0 0 18px var(--cube-glow-local);
}

.cube--local .face-top { opacity: 0.95; }
.cube--local .face-front { filter: brightness(0.8); }
.cube--local .face-right { filter: brightness(0.6); }

/* Empty State - Ghostly Wireframe */
.cube--empty .face {
  background: var(--cube-face-empty);
  border: 1px solid rgba(255, 255, 255, 0.1);
}

.cube--empty .face-top { opacity: 0.3; }
.cube--empty .face-front { opacity: 0.2; }
.cube--empty .face-right { opacity: 0.1; }

.cube--hovered {
  transform: scale(1.1) translateZ(10px) !important;
  transition: transform 0.2s ease-out;
}

/* Tooltip Styles */
.cube-tooltip {
  position: absolute;
  bottom: 100%;
  left: 50%;
  transform: translateX(-50%) rotateZ(45deg) rotateX(-54deg) translateY(-20px);
  background: rgba(0, 0, 0, 0.85);
  backdrop-filter: blur(4px);
  padding: var(--space-2) var(--space-3);
  border-radius: var(--radius-sm);
  border: 1px solid rgba(255, 255, 255, 0.1);
  white-space: nowrap;
  pointer-events: none;
  z-index: 100;
  display: flex;
  flex-direction: column;
  gap: 2px;
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.3);
}

.tooltip-name {
  color: #fff;
  font-size: 11px;
  font-weight: 600;
  font-family: var(--font-mono);
}

.tooltip-status {
  display: flex;
  align-items: center;
  gap: var(--space-1);
  font-size: 9px;
  font-weight: 700;
  letter-spacing: 0.05em;
}

.status-dot {
  width: 5px;
  height: 5px;
  border-radius: 50%;
}

.tooltip-status.online { color: var(--color-success); }
.tooltip-status.online .status-dot { background: var(--color-success); }

.tooltip-status.offline { color: var(--color-error); }
.tooltip-status.offline .status-dot { background: var(--color-error); }

.tooltip-status.local { color: var(--color-accent); }
.tooltip-status.local .status-dot { background: var(--color-accent); }

/* Transitions */
.fade-enter-active,
.fade-leave-active {
  transition: opacity 0.2s ease, transform 0.2s ease;
}

.fade-enter-from,
.fade-leave-to {
  opacity: 0;
  transform: translateX(-50%) rotateZ(45deg) rotateX(-54deg) translateY(-10px);
}

@keyframes cube-float {
  0%, 100% { transform: translateZ(0); }
  25% { transform: translateZ(12px); }
  50% { transform: translateZ(2px); }
  75% { transform: translateZ(-4px); }
}

@media (max-width: 767px) {
  .cube-scene {
    min-width: 240px;
    min-height: 260px;
    padding: var(--space-4);
  }

  .cube-viewport {
    width: 240px;
    height: 240px;
    perspective: 800px;
  }

  .cube-grid {
    width: 120px;
    height: 120px;
  }

  .cube { width: 36px; height: 36px; }
  .face { width: 36px; height: 36px; }
  .face-top { transform: rotateX(90deg) translateZ(18px); }
  .face-front { transform: translateZ(18px); }
  .face-right { transform: rotateY(90deg) translateZ(18px); }
}
</style>
