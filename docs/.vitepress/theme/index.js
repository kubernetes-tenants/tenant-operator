import DefaultTheme from 'vitepress/theme';
import AnimatedDiagram from '../components/AnimatedDiagram.vue';
import HowItWorksDiagram from '../components/HowItWorksDiagram.vue';
import ComparisonSection from '../components/ComparisonSection.vue';
import './custom.css';

export default {
  extends: DefaultTheme,
  enhanceApp({ app }) {
    // Register custom components globally
    app.component('AnimatedDiagram', AnimatedDiagram);
    app.component('HowItWorksDiagram', HowItWorksDiagram);
    app.component('ComparisonSection', ComparisonSection);
  }
};
