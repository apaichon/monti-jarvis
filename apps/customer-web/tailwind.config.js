/** @type {import('tailwindcss').Config} */
export default {
  content: ['./src/**/*.{html,js,svelte,ts}'],
  theme: {
    extend: {
      colors: {
        monti: {
          bg: '#020712',
          panel: '#08172a',
          line: '#3d6ea0',
          cyan: '#00b7ff',
          blue: '#006dff'
        }
      },
      boxShadow: {
        neon: '0 0 36px rgb(0 132 255 / 36%)'
      }
    }
  },
  plugins: []
};