/**
 * Skill 管理器 Mixin - 管理 skill 预览面板的流式状态
 */
import Vue from 'vue';

// 使用 Vue.observable 创建响应式共享状态
const sharedState = Vue.observable({
  // 预览面板的流式状态
  previewIsStreaming: false,
  previewStreamingIds: [],
  // 主会话的流式状态
  mainIsStreaming: false,
  mainStreamingThreadIds: [],
});

export default {
  computed: {
    // 预览面板的流式状态（响应式读取共享状态）
    previewIsStreaming: {
      get() {
        return sharedState.previewIsStreaming;
      },
      set(value) {
        sharedState.previewIsStreaming = value;
      },
    },
    previewStreamingIds: {
      get() {
        return sharedState.previewStreamingIds;
      },
      set(value) {
        sharedState.previewStreamingIds = value;
      },
    },
    // 主会话的流式状态（响应式读取共享状态）
    mainIsStreaming: {
      get() {
        return sharedState.mainIsStreaming;
      },
      set(value) {
        sharedState.mainIsStreaming = value;
      },
    },
    mainStreamingThreadIds: {
      get() {
        return sharedState.mainStreamingThreadIds;
      },
      set(value) {
        sharedState.mainStreamingThreadIds = value;
      },
    },
  },
};
