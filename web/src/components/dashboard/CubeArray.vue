<template>
  <div class="cube-scene" role="img" aria-label="3D cube grid representing machine connectivity status">
    <div class="cube-viewport">
      <div class="cube-grid">
        <div
          v-for="(cell, i) in displayCubes"
          :key="i"
          class="cube-cell"
          :style="cellPositionStyle(i)"
          @mouseenter="onCellMouseEnter($event, i)"
          @mouseleave="onCellMouseLeave"
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
        </div>
      </div>
    </div>

    <!-- Global Tooltip (outside 3D context via Teleport) -->
    <Teleport to="body">
      <Transition name="fade">
        <div
          v-if="tooltipData"
          class="cube-tooltip"
          :style="tooltipStyle"
        >
          <div class="tooltip-name">{{ tooltipData.name }}</div>
          <div class="tooltip-status" :class="tooltipData.status">
            <span class="status-dot"></span>
            {{ tooltipData.status.toUpperCase() }}
          </div>
        </div>
      </Transition>
    </Teleport>
  </div>
</template>

<script setup lang="ts">
import { computed, ref } from 'vue'
import { t } from '@/i18n'

export interface CubeData {
  id?: string
  name?: string
  status: 'online' | 'offline' | 'empty' | 'local' | 'unknown'
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
const tooltipStyle = ref<Record<string, string>>({})
const tooltipData = ref<{ name: string; status: string } | null>(null)

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

function getCellName(cell: CubeData): string {
  if (cell.status === 'empty') return ''
  return cell.name || (cell.status === 'local' ? t('common.local') : t('common.unknown'))
}

function onCellMouseEnter(event: MouseEvent, index: number) {
  const cell = displayCubes.value[index]
  if (!cell || cell.status === 'empty') return

  hoveredIndex.value = index
  tooltipData.value = {
    name: getCellName(cell),
    status: cell.status
  }

  const el = event.currentTarget as HTMLElement
  const rect = el.getBoundingClientRect()
  tooltipStyle.value = {
    position: 'fixed',
    left: `${rect.left + rect.width / 2}px`,
    top: `${rect.top}px`,
    transform: 'translate(-50%, -100%)',
    marginTop: '-8px'
  }
}

function onCellMouseLeave() {
  hoveredIndex.value = null
  tooltipData.value = null
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

/* Unknown State - Yellow Glow (connection not verified) */
.cube--unknown .face {
  background: var(--cube-face-unknown);
  box-shadow: 0 0 15px var(--cube-glow-unknown);
}

.cube--unknown .face-top { opacity: 0.95; }
.cube--unknown .face-front { filter: brightness(0.8); }
.cube--unknown .face-right { filter: brightness(0.6); }

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

<!-- Tooltip is Teleported to body, so global styles apply -->
<style>
.cube-tooltip {
  position: fixed;
  z-index: 9999;
  pointer-events: none;
  background: rgba(0, 0, 0, 0.88);
  backdrop-filter: blur(6px);
  padding: 6px 10px;
  border-radius: 4px;
  border: 1px solid rgba(255, 255, 255, 0.1);
  white-space: nowrap;
  display: flex;
  flex-direction: column;
  gap: 2px;
  box-shadow: 0 4px 16px rgba(0, 0, 0, 0.4);
}

.tooltip-name {
  color: #fff;
  font-size: 11px;
  font-weight: 600;
  font-family: var(--font-mono), monospace;
}

.tooltip-status {
  display: flex;
  align-items: center;
  gap: 4px;
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
  transition: opacity 0.15s ease, transform 0.15s ease;
}

.fade-enter-from,
.fade-leave-to {
  opacity: 0;
  transform: translateY(4px);
}
</style>
