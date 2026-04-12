import type { Config } from 'tailwindcss'

export default {
  content: ['./index.html', './src/**/*.{ts,tsx}'],
  theme: {
    extend: {
      colors: {
        cg: {
          bg: '#0a0c10',
          panel: '#111318',
          border: '#1e2028',
          accent: '#00ff88',
          'accent-dim': '#00cc6a',
          muted: '#6b7280',
          danger: '#ef4444',
        },
      },
      fontFamily: {
        mono: ['JetBrains Mono', 'Fira Code', 'monospace'],
        display: ['Space Grotesk', 'system-ui', 'sans-serif'],
      },
    },
  },
  plugins: [],
} satisfies Config
