import DefaultTheme from 'vitepress/theme';
import Layout from './Layout.vue';
import AnimatedDiagram from '../components/AnimatedDiagram.vue';
import HowItWorksDiagram from '../components/HowItWorksDiagram.vue';
import ComparisonSection from '../components/ComparisonSection.vue';
import DependencyGraphVisualizer from '../components/DependencyGraphVisualizer.vue';
import TemplateBuilder from '../components/TemplateBuilder.vue';
import AnnouncementBanner from '../components/AnnouncementBanner.vue';
import QuickstartStep from '../components/QuickstartStep.vue';
import CreationPolicyVisualizer from '../components/CreationPolicyVisualizer.vue';
import DeletionPolicyVisualizer from '../components/DeletionPolicyVisualizer.vue';
import ConflictPolicyVisualizer from '../components/ConflictPolicyVisualizer.vue';
import PatchStrategyVisualizer from '../components/PatchStrategyVisualizer.vue';
import './custom.css';

export default {
  extends: DefaultTheme,
  Layout,
  enhanceApp({ app }) {
    // Register custom components globally
    app.component('AnimatedDiagram', AnimatedDiagram);
    app.component('HowItWorksDiagram', HowItWorksDiagram);
    app.component('ComparisonSection', ComparisonSection);
    app.component('DependencyGraphVisualizer', DependencyGraphVisualizer);
    app.component('TemplateBuilder', TemplateBuilder);
    app.component('AnnouncementBanner', AnnouncementBanner);
    app.component('QuickstartStep', QuickstartStep);
    app.component('CreationPolicyVisualizer', CreationPolicyVisualizer);
    app.component('DeletionPolicyVisualizer', DeletionPolicyVisualizer);
    app.component('ConflictPolicyVisualizer', ConflictPolicyVisualizer);
    app.component('PatchStrategyVisualizer', PatchStrategyVisualizer);
  }
};
