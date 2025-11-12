import DefaultTheme from 'vitepress/theme';
import AnimatedDiagram from '../components/AnimatedDiagram.vue';
import './custom.css';

export default {
  extends: DefaultTheme,
  enhanceApp({ app }) {
    // Register custom components globally
    app.component('AnimatedDiagram', AnimatedDiagram);
  }
};
