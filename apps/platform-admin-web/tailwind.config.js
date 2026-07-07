/** @type {import('tailwindcss').Config} */
export default {
  content: ['./src/**/*.{html,js,svelte,ts}'],
  theme: {
    extend: {
      colors: {
        admin: {
          bg: '#0a0f1a',
          panel: '#0c1424',
          line: 'rgb(0 183 255 / 18%)',
          cyan: '#00b7ff',
          danger: '#ff5c7a',
          success: '#3dd68c',
          muted: '#8fa5bf'
        }
      }
    }
  },
  plugins: []
};