/** @type {import('tailwindcss').Config} */
export default {
  content: ['./index.html', './src/**/*.{svelte,js,ts}'],
  darkMode: 'class',
  theme: {
    extend: {
      colors: {
        accent: {
          DEFAULT: '#4F7CFF',
          hover: '#3D68E8',
          light: '#EEF3FF',
          dark: '#1E294B',
        },
      },
      borderRadius: {
        DEFAULT: '12px',
        card: '12px',
      },
      transitionDuration: {
        DEFAULT: '150ms',
      },
    },
  },
  plugins: [],
};
