<template>
  <div class="general-agent-page">
    <!-- 左侧会话列表 - 可折叠 -->
    <div :class="['sidebar', { collapsed: sidebarCollapsed }]">
      <div class="sidebar-content">
        <div class="sidebar-header">
          <el-button
            type="primary"
            class="new-chat-btn"
            @click="initNewConversation"
          >
            <i class="el-icon-plus"></i>
            新建对话
          </el-button>
        </div>

        <div class="sidebar-divider"></div>

        <div class="conversation-list">
          <div
            v-for="item in conversationList"
            :key="item.threadId"
            :class="[
              'conversation-item',
              { active: currentThreadId === item.threadId },
            ]"
            @click="selectConversation(item.threadId)"
          >
            <i class="el-icon-chat-dot-round"></i>
            <span class="conversation-title">{{ item.title || '新对话' }}</span>
            <el-dropdown trigger="click" @command="handleCommand($event, item)">
              <i class="el-icon-more" @click.stop></i>
              <el-dropdown-menu slot="dropdown">
                <el-dropdown-item command="delete">
                  <span style="color: #f56c6c">删除</span>
                </el-dropdown-item>
              </el-dropdown-menu>
            </el-dropdown>
          </div>
        </div>
      </div>
    </div>

    <!-- 主内容区 -->
    <div
      class="agent-main-content"
      :class="{ 'has-workspace': panelVisible && activeWorkspace }"
    >
      <!-- 主消息区域 -->
      <div class="main-content-body">
        <!-- 顶部标题栏 -->
        <div class="header">
          <div class="header-left">
            <button class="sidebar-toggle-btn" @click="toggleSidebar">
              <i
                :class="
                  sidebarCollapsed ? 'el-icon-s-unfold' : 'el-icon-s-fold'
                "
              ></i>
            </button>
            <div class="header-title">{{ currentTitle }}</div>
          </div>
        </div>

        <!-- 消息区域 - 独立滚动 -->
        <div
          :class="[
            'message-area',
            { empty: isEmptyConversation && !isLoadingHistory },
          ]"
          ref="messageArea"
          @scroll="handleMessageAreaScroll"
        >
          <!-- 加载历史记录中 -->
          <div v-if="isLoadingHistory" class="history-loading">
            <i class="el-icon-loading"></i>
            <span>加载中...</span>
          </div>

          <!-- 消息列表 -->
          <div
            v-else-if="messageList.length > 0 || isStreaming"
            class="message-list"
          >
            <message-item
              v-for="(msg, index) in messageList"
              :key="msg.id || index"
              :message="msg"
              :tool-results="getToolResultsForMessage(msg)"
              :is-last-message="index === messageList.length - 1"
              :thread-id="currentThreadId"
              @regenerate="handleRegenerate"
              @view-workspace="handleViewWorkspace"
            />
          </div>

          <div ref="scrollAnchor"></div>
        </div>

        <!-- 滚动到底部按钮 -->
        <transition name="scroll-btn-fade">
          <button
            v-if="showScrollToBottom"
            class="scroll-to-bottom-btn"
            @click="handleScrollToBottomClick"
          >
            <svg
              viewBox="0 0 24 24"
              width="16"
              height="16"
              fill="none"
              stroke="currentColor"
              stroke-width="2"
              stroke-linecap="round"
              stroke-linejoin="round"
            >
              <polyline points="6,9 12,15 18,9"></polyline>
            </svg>
          </button>
        </transition>

        <!-- 底部输入区 -->
        <div
          :class="[
            'input-area',
            { 'is-centered': isEmptyConversation && !isLoadingHistory },
          ]"
        >
          <!-- 欢迎词 - 仅居中时显示 -->
          <div
            v-if="isEmptyConversation && !isLoadingHistory"
            class="welcome-section"
          >
            <div class="welcome-avatar">
              <img
                v-if="assistantAvatar"
                :src="assistantAvatar"
                alt="Assistant"
              />
              <i v-else class="el-icon-cpu"></i>
            </div>
            <div class="welcome-title">你好，我是万悟</div>
          </div>

          <div class="input-container">
            <!-- 文件预览 -->
            <div v-if="uploadedFiles.length > 0" class="file-preview">
              <div
                v-for="(file, index) in uploadedFiles"
                :key="index"
                class="file-item"
                :class="{ 'is-uploading': file.uploading }"
              >
                <img
                  v-if="file.type.startsWith('image/')"
                  :src="file.displayUrl || file.url"
                  class="file-thumb"
                />
                <div v-else class="file-icon">
                  <i class="el-icon-document"></i>
                </div>
                <!-- 上传进度遮罩 -->
                <div v-if="file.uploading" class="upload-overlay">
                  <div class="upload-progress-bar">
                    <svg viewBox="0 0 36 36" width="36" height="36">
                      <circle class="progress-bg" cx="18" cy="18" r="15" />
                      <circle
                        class="progress-fill"
                        cx="18"
                        cy="18"
                        r="15"
                        :stroke-dasharray="94.2"
                        :stroke-dashoffset="
                          94.2 - (94.2 * (file.uploadProgress || 0)) / 100
                        "
                      />
                    </svg>
                    <span class="progress-text">
                      {{ file.uploadProgress || 0 }}
                    </span>
                  </div>
                </div>
                <i
                  class="el-icon-close file-remove"
                  @click="removeFile(index)"
                ></i>
              </div>
            </div>

            <!-- 输入框 -->
            <div class="input-wrapper">
              <el-input
                v-model="inputMessage"
                type="textarea"
                :rows="1"
                :autosize="{ minRows: 1, maxRows: 6 }"
                placeholder="输入问题，按 Enter 发送，Shift+Enter 换行"
                @keydown.enter.native="handleKeyDown"
                :disabled="isStreaming"
              />
            </div>

            <!-- 底部工具栏：模型选择 + 发送按钮 -->
            <div class="input-toolbar">
              <div class="toolbar-left">
                <ModelSelect
                  v-model="selectedModel"
                  :options="modelList"
                  placeholder="选择模型"
                  :loading="modelLoading"
                  :filterable="true"
                  @change="handleModelChange"
                  class="model-select-inline"
                />
                <div
                  class="config-btn"
                  :class="{ 'has-selection': selectedTools.length > 0 }"
                  @click="showConfigDrawer = true"
                >
                  <i class="el-icon-setting"></i>
                  <span>配置</span>
                  <el-badge
                    v-if="selectedTools.length > 0"
                    :value="selectedTools.length"
                    type="primary"
                  />
                </div>
              </div>
              <div class="toolbar-right">
                <el-upload
                  action="#"
                  :auto-upload="false"
                  :show-file-list="false"
                  :on-change="handleFileChange"
                  multiple
                >
                  <el-tooltip content="上传文件" placement="top">
                    <i class="action-icon el-icon-paperclip"></i>
                  </el-tooltip>
                </el-upload>
                <el-button
                  v-if="isStreaming"
                  class="send-btn stop-btn"
                  circle
                  @click="stopStreaming"
                >
                  <svg
                    class="stop-icon"
                    viewBox="0 0 24 24"
                    width="16"
                    height="16"
                  >
                    <rect x="6" y="6" width="12" height="12" rx="2" />
                  </svg>
                </el-button>
                <el-button
                  v-else
                  type="primary"
                  class="send-btn"
                  circle
                  :disabled="!canSend"
                  @click="sendMessage"
                >
                  <svg
                    class="send-icon"
                    viewBox="0 0 24 24"
                    width="18"
                    height="18"
                  >
                    <path
                      fill="currentColor"
                      d="M2.01 21L23 12 2.01 3 2 10l15 2-15 2z"
                    />
                  </svg>
                </el-button>
              </div>
            </div>
          </div>
          <div v-if="!isEmptyConversation" class="input-footer">
            <span>通用智能体 · 内容由 AI 生成，仅供参考</span>
          </div>
        </div>
      </div>

      <!-- Workspace 面板 -->
      <transition name="workspace-slide">
        <workspace-panel
          v-if="panelVisible && activeWorkspace"
          ref="workspacePanel"
          :thread-id="activeWorkspace.threadId"
          :run-id="activeWorkspace.runId"
          :initial-data="currentWorkspaceTree"
          @close="hidePanel"
          @preview-file="handlePreviewFile"
        />
      </transition>

      <!-- 文件预览抽屉 -->
      <file-preview-drawer
        :visible.sync="previewVisible"
        :file="previewFile"
        :file-path="previewFilePath"
        :file-ext="previewFileExt"
        :type="previewType"
        :url="previewUrl"
        :content="previewContent"
        :loading="previewLoading"
        :panel-style="previewPanelStyle"
        :excel-data="previewExcelData"
        @download="downloadPreviewFile"
        @close="closePreview"
      />

      <!-- 配置抽屉 -->
      <config-drawer
        :visible.sync="showConfigDrawer"
        :tool-list="toolList"
        :selected-tools="selectedTools"
        :loading="loadingTools"
        @toggle-tool="toggleTool"
        @close="showConfigDrawer = false"
      />
    </div>
  </div>
</template>

<script>
import MessageItem from './components/MessageItem.vue';
import WorkspacePanel from './components/WorkspacePanel.vue';
import FilePreviewDrawer from './components/FilePreviewDrawer.vue';
import ConfigDrawer from './components/ConfigDrawer.vue';
import ModelSelect from '@/components/modelSelect.vue';
import {
  getGeneralAgentConversationList,
  createGeneralAgentConversation,
  deleteGeneralAgentConversation,
  getGeneralAgentConversationDetail,
  getGeneralAgentConfig,
  updateGeneralAgentConfig,
  chatGeneralAgentConversation,
  getGeneralAgentToolSelect,
  getLlmModelSelect,
  getGeneralAgentWorkspace,
  uploadGeneralAgentFile,
  previewGeneralAgentWorkspace,
  downloadGeneralAgentWorkspace,
} from '@/api/generalAgent';
import {
  SSEEventParser,
  EventType,
  ActivityType,
  ActivityStatus,
} from './utils/sse-parser';
import { formatDuration } from './utils/helpers';
import { mapState, mapActions, mapGetters } from 'vuex';
import { avatarSrc } from '@/utils/util';
import * as XLSX from 'xlsx';

export default {
  name: 'GeneralAgent',
  components: {
    MessageItem,
    WorkspacePanel,
    FilePreviewDrawer,
    ConfigDrawer,
    ModelSelect,
  },
  data() {
    return {
      sidebarCollapsed: true,
      conversationList: [],
      currentThreadId: '',
      pageNo: 1,
      pageSize: 50,
      total: 0,
      isNewConversation: false,
      isLoadingHistory: false,

      // 每个会话独立的消息列表 { threadId: messageList }
      messagesMap: {},
      inputMessage: '',
      uploadedFiles: [],
      // 每个会话独立的流式状态 { threadId: { isStreaming, abortController, streamingMessage } }
      streamingMap: {},

      selectedModel: '',
      selectedTools: [],
      selectedAssistants: [],
      modelList: [],
      modelLoading: false,
      toolList: [],
      loadingTools: false,
      showConfigDrawer: false,

      currentRunId: '',
      currentStage: '',

      // Workspace 相关
      workspacePanelVisible: false,
      workspaceLoading: false,
      workspaceInfo: null,

      // 文件预览
      previewVisible: false,
      previewLoading: false,
      previewFile: null,
      previewFilePath: '',
      previewFileExt: '',
      previewUrl: '',
      previewContent: '',
      previewType: '',
      previewBlobUrl: '',
      previewExcelData: null,
      workspaceRect: null,
      resizeObserver: null,

      // 滚动控制
      userHasScrolled: false,
      showScrollToBottom: false,
      isAutoScrolling: false,
    };
  },
  computed: {
    ...mapState('workspace', ['activeWorkspace', 'panelVisible']),
    ...mapGetters('workspace', ['hasWorkspace', 'currentWorkspaceTree']),
    ...mapGetters('user', ['commonInfo']),

    assistantAvatar() {
      const tab = this.commonInfo?.data?.tab || {};
      const path = tab.logo?.path;
      return path ? avatarSrc(path) : null;
    },

    // 当前会话的消息列表
    messageList: {
      get() {
        return this.messagesMap[this.currentThreadId] || [];
      },
      set(val) {
        this.$set(this.messagesMap, this.currentThreadId, val);
      },
    },
    // 当前会话的流式状态
    currentStreaming() {
      return (
        this.streamingMap[this.currentThreadId] || {
          isStreaming: false,
          abortController: null,
          streamingMessage: null,
        }
      );
    },
    isStreaming() {
      return this.currentStreaming.isStreaming;
    },
    streamingMessage() {
      return this.currentStreaming.streamingMessage;
    },

    currentTitle() {
      if (!this.currentThreadId) return '';
      const conv = this.conversationList.find(
        c => c.threadId === this.currentThreadId,
      );
      return conv?.title || '新对话';
    },
    canSend() {
      const hasContent =
        this.inputMessage.trim() || this.uploadedFiles.length > 0;
      const hasModel = !!this.selectedModel;
      return hasContent && hasModel;
    },
    previewPanelStyle() {
      // 计算宽度：屏幕宽度的一半，最小 500px
      const screenWidth = window.screen.width;
      const halfScreenWidth = Math.floor(screenWidth / 2);
      const width = Math.max(500, halfScreenWidth);

      if (!this.workspaceRect) {
        // 默认情况：workspace 未显示时，使用计算值
        const sidebarWidth = this.sidebarCollapsed ? 0 : 240;
        const sidebarMargin = this.sidebarCollapsed ? 0 : 16;
        const workspaceWidth = 400;
        const pagePadding = 16;
        return {
          right: `${pagePadding + sidebarWidth + sidebarMargin + workspaceWidth}px`,
          width: `${width}px`,
        };
      }
      // workspace 显示时，紧贴其左边缘
      const rightEdge = window.innerWidth - this.workspaceRect.left;
      return {
        right: `${rightEdge}px`,
        width: `${width}px`,
      };
    },
    hasAssistantContent() {
      return this.messageList.some(
        m =>
          m.role === 'assistant' &&
          (m.content || m.reasoning || (m.toolCalls && m.toolCalls.length > 0)),
      );
    },
    isEmptyConversation() {
      return this.messageList.length === 0;
    },
    // Workspace 相关
    workspaceThreadAndRun() {
      if (this.activeWorkspace && this.currentThreadId) {
        return {
          threadId: this.currentThreadId,
          runId: this.activeWorkspace.runId,
        };
      }
      return null;
    },
  },
  watch: {
    panelVisible(val) {
      this.workspacePanelVisible = val;
      if (val && this.activeWorkspace) {
        this.loadWorkspaceFiles();
        this.$nextTick(() => this.updateWorkspaceRect());
      } else if (!val) {
        // 工作空间关闭时，关闭文件预览
        this.previewVisible = false;
        this.closePreview();
      }
    },
    previewVisible(val) {
      if (val) {
        this.$nextTick(() => this.updateWorkspaceRect());
      }
    },
  },
  mounted() {
    this.initNewConversation();
    this.fetchModelList();
    this.fetchConversationList();
    this.fetchToolList();
    this.initUserInfo();
    this.setupResizeObserver();
  },
  beforeDestroy() {
    // 清理所有会话的流式状态
    Object.keys(this.streamingMap).forEach(threadId => {
      const streaming = this.streamingMap[threadId];
      if (streaming && streaming.abortController) {
        streaming.abortController.abort();
      }
    });
    this.streamingMap = {};
    this.reset();
    if (this.resizeObserver) {
      this.resizeObserver.disconnect();
      this.resizeObserver = null;
    }
  },
  methods: {
    ...mapActions('workspace', [
      'handleWorkspaceActivity',
      'showPanel',
      'hidePanel',
      'setWorkspaceTree',
      'setActiveWorkspace',
      'clearWorkspace',
      'reset',
    ]),
    ...mapActions('user', ['getPermissionInfo', 'getCommonInfo']),

    async initUserInfo() {
      if (localStorage.getItem('access_cert')) {
        await this.getPermissionInfo();
        await this.getCommonInfo();
      }
    },

    setupResizeObserver() {
      if (typeof ResizeObserver === 'undefined') return;
      this.resizeObserver = new ResizeObserver(() => {
        this.updateWorkspaceRect();
      });
      // 监听整个页面容器
      const pageEl = this.$el;
      if (pageEl) {
        this.resizeObserver.observe(pageEl);
      }
      // 也监听 sidebar 变化
      const sidebar = pageEl?.querySelector('.sidebar');
      if (sidebar) {
        this.resizeObserver.observe(sidebar);
      }
    },

    updateWorkspaceRect() {
      this.$nextTick(() => {
        const workspaceEl = this.$refs.workspacePanel?.$el;
        if (workspaceEl) {
          this.workspaceRect = workspaceEl.getBoundingClientRect();
        } else {
          // 如果 workspace 不可见，使用 mainContent 的右边界
          const mainContent = this.$el?.querySelector('.agent-main-content');
          if (mainContent) {
            const rect = mainContent.getBoundingClientRect();
            this.workspaceRect = { left: rect.right };
          }
        }
      });
    },

    toggleSidebar() {
      this.sidebarCollapsed = !this.sidebarCollapsed;
    },

    async fetchModelList() {
      this.modelLoading = true;
      try {
        const res = await getLlmModelSelect();
        if (res.code === 0 && res.data?.list) {
          this.modelList = res.data.list.map(model => ({
            modelId: model.modelId || model.model,
            displayName: model.displayName || model.modelName || model.model,
            modelName: model.displayName || model.modelName || model.model,
            model: model.model,
            provider: model.provider,
            modelType: model.modelType,
            config: model.config,
            avatar: model.avatar || { path: '' },
            tags: model.tags || [],
          }));
        }
      } catch (error) {
        console.error('获取模型列表失败:', error);
      } finally {
        this.modelLoading = false;
      }
    },

    async fetchConversationList() {
      try {
        const res = await getGeneralAgentConversationList({
          pageNo: this.pageNo,
          pageSize: this.pageSize,
        });
        if (res.code === 0) {
          this.conversationList = res.data?.list || [];
          this.total = res.data?.total || 0;
        }
      } catch (error) {
        console.error('获取对话列表失败:', error);
      }
    },

    async fetchToolList() {
      this.loadingTools = true;
      try {
        const res = await getGeneralAgentToolSelect();
        if (res.code === 0 && res.data) {
          this.toolList = res.data || [];
        }
      } catch (error) {
        console.error('获取工具列表失败:', error);
      } finally {
        this.loadingTools = false;
      }
    },

    initNewConversation() {
      this.currentThreadId = '';
      this.isNewConversation = true;
      this.$set(this.messagesMap, '', []);
      this.selectedTools = [];
      // 重置滚动状态
      this.userHasScrolled = false;
      this.showScrollToBottom = false;
      // 关闭工作区面板
      this.hidePanel();
      if (this.modelList && this.modelList.length > 0) {
        const defaultModel = this.modelList[0];
        this.selectedModel = defaultModel?.modelId || '';
      }
    },

    async createConversationWithTitle(title) {
      try {
        if (!this.modelList || this.modelList.length === 0) {
          this.$message.warning('模型列表加载中，请稍后重试');
          return null;
        }

        // 使用用户选择的模型，如果没有选择则使用第一个模型
        const selectedModelConfig = this.selectedModel
          ? this.modelList.find(m => m.modelId === this.selectedModel)
          : this.modelList[0];

        const modelConfig = {
          modelId: selectedModelConfig?.modelId || '',
          model: selectedModelConfig?.model || '',
          provider: selectedModelConfig?.provider || '',
          displayName: selectedModelConfig?.modelName || '',
          modelType: selectedModelConfig?.modelType || 'llm',
          config: selectedModelConfig?.config || {},
        };

        const res = await createGeneralAgentConversation({
          title: title || '新对话',
          modelConfig,
        });

        if (res.code === 0) {
          const threadId = res.data?.threadId;
          if (threadId) {
            this.currentThreadId = threadId;
            this.isNewConversation = false;

            const oldMessages = this.messagesMap[''] || [];
            this.$set(this.messagesMap, threadId, oldMessages);
            this.$delete(this.messagesMap, '');

            this.selectedModel = modelConfig.modelId;
            this.selectedTools = [];
            this.conversationList.unshift({
              threadId,
              title: title || '新对话',
              createdAt: new Date().toISOString(),
            });
            return threadId;
          } else {
            this.$message.error('创建对话失败：未返回对话ID');
          }
        } else {
          this.$message.error(res.msg || '创建对话失败');
        }
        return null;
      } catch (error) {
        console.error('创建对话失败:', error);
        this.$message.error('创建对话失败，请检查网络连接');
        return null;
      }
    },

    async createConversation() {
      try {
        // 检查模型列表是否已加载
        if (!this.modelList || this.modelList.length === 0) {
          this.$message.warning('模型列表加载中，请稍后重试');
          return false;
        }

        // 获取默认模型配置
        const defaultModel = this.modelList[0];
        const modelConfig = {
          modelId: defaultModel?.modelId || '',
          model: defaultModel?.model || '',
          provider: defaultModel?.provider || '',
          displayName: defaultModel?.modelName || '',
          modelType: defaultModel?.modelType || 'llm',
          config: defaultModel?.config || {},
        };

        const res = await createGeneralAgentConversation({
          title: '新对话',
          modelConfig,
        });
        if (res.code === 0) {
          const threadId = res.data?.threadId;
          if (threadId) {
            this.currentThreadId = threadId;
            // 初始化新会话的消息列表
            this.$set(this.messagesMap, threadId, []);
            this.selectedModel = modelConfig.modelId;
            this.selectedTools = [];
            this.conversationList.unshift({
              threadId,
              title: '新对话',
              createdAt: new Date().toISOString(),
            });
            await new Promise(resolve => setTimeout(resolve, 500));
            return true;
          } else {
            this.$message.error('创建对话失败：未返回对话ID');
          }
        } else {
          this.$message.error(res.msg || '创建对话失败');
        }
        return false;
      } catch (error) {
        console.error('创建对话失败:', error);
        this.$message.error('创建对话失败，请检查网络连接');
        return false;
      }
    },

    selectConversation(threadId) {
      if (this.currentThreadId === threadId) return;
      // 切换会话时，只切换 currentThreadId，不中止 SSE 流
      // SSE 流会继续在后台运行，切换回来时能继续显示
      this.currentThreadId = threadId;
      this.isNewConversation = false;
      this.isLoadingHistory = true;
      // 重置滚动状态
      this.userHasScrolled = false;
      this.showScrollToBottom = false;
      // 关闭工作区面板
      this.hidePanel();
      // 清空当前会话的消息缓存，确保每次都重新请求
      this.$set(this.messagesMap, threadId, []);
      this.fetchHistory();
    },

    async fetchHistory() {
      if (!this.currentThreadId) return;

      try {
        const res = await getGeneralAgentConversationDetail({
          threadId: this.currentThreadId,
          pageNo: 1,
          pageSize: 100,
        });

        if (res.code === 0 && res.data?.list) {
          const allMessages = [];
          res.data.list.forEach(run => {
            // 后端返回的是 events 字段，需要聚合为消息
            if (run.events && Array.isArray(run.events)) {
              const messages = this.aggregateEventsToMessages(run.events);
              allMessages.push(...messages);
            }
            // 兼容旧格式 messages 字段
            if (run.messages && Array.isArray(run.messages)) {
              run.messages.forEach(msg => {
                const formatted = this.formatMessage(msg);
                if (formatted) {
                  allMessages.push(formatted);
                }
              });
            }
            if (run.runId) this.currentRunId = run.runId;
          });
          // 使用 $set 确保响应式
          this.$set(this.messagesMap, this.currentThreadId, allMessages);
          // 先关闭加载状态，让消息列表渲染
          this.isLoadingHistory = false;
          // 等待 DOM 渲染完成后滚动到底部
          this.$nextTick(() => {
            requestAnimationFrame(() => {
              this.scrollToBottom(true);
            });
          });
        } else {
          this.isLoadingHistory = false;
        }
        this.loadConfig();
      } catch (error) {
        console.error('获取历史消息失败:', error);
        this.isLoadingHistory = false;
      }
    },

    // 将 AG-UI 事件聚合为消息 - 支持交错展示和 activity 嵌套
    aggregateEventsToMessages(events) {
      const messages = [];
      const toolCallMap = new Map();
      const activityStack = []; // 用于跟踪嵌套的 activity
      let currentActivity = null; // 当前的 activity

      const getCurrentActivity = () => currentActivity;

      const addFragment = fragment => {
        if (currentActivity) {
          currentActivity.fragments.push(fragment);
        } else {
          messages.push({
            ...fragment,
            id: fragment.id || this.generateId(),
            role: 'assistant',
            timestamp: Date.now(),
          });
        }
      };

      for (const event of events) {
        const eventTimestamp = event.timestamp
          ? new Date(event.timestamp).getTime()
          : Date.now();

        switch (event.type) {
          case 'RUN_STARTED': {
            if (event.input?.messages && Array.isArray(event.input.messages)) {
              event.input.messages.forEach(msg => {
                if (msg.role === 'user') {
                  messages.push({
                    id: msg.id || this.generateId(),
                    role: 'user',
                    content: this.formatContent(msg.content),
                    files: this.extractFilesFromContent(msg.content),
                    toolCalls: null,
                    toolResults: null,
                    toolCallId: null,
                    reasoning: '',
                    timestamp: eventTimestamp,
                  });
                }
              });
            }
            break;
          }

          case 'ACTIVITY_SNAPSHOT': {
            const activityContent = event.content || {};
            if (event.activityType === 'sub_agent') {
              if (activityContent.status === 'started') {
                currentActivity = {
                  type: 'activity',
                  activityType: event.activityType,
                  activityId: event.activityId || '',
                  agentName: activityContent.agentName || '',
                  fragments: [],
                };
                activityStack.push(currentActivity);
              } else if (activityContent.status === 'finished') {
                if (activityStack.length > 0) {
                  const finishedActivity = activityStack.pop();
                  if (activityStack.length > 0) {
                    currentActivity = activityStack[activityStack.length - 1];
                    currentActivity.fragments.push(finishedActivity);
                  } else {
                    messages.push({
                      id: event.messageId || this.generateId(),
                      role: 'assistant',
                      ...finishedActivity,
                      timestamp: eventTimestamp,
                    });
                    currentActivity = null;
                  }
                }
              }
            } else if (
              event.activityType === 'workspace' &&
              activityContent.runId
            ) {
              this.handleWorkspaceActivity({
                runId: activityContent.runId,
                threadId: activityContent.threadId || this.currentThreadId,
                fileCount: activityContent.fileCount || 0,
                totalSize: activityContent.totalSize || 0,
                timestamp: activityContent.timestamp || eventTimestamp,
              });

              addFragment({
                type: 'workspace',
                workspaceInfo: {
                  fileCount: activityContent.fileCount || 0,
                  totalSize: activityContent.totalSize || 0,
                },
                runId: activityContent.runId,
              });
            }
            break;
          }

          case 'REASONING_MESSAGE_START': {
            addFragment({
              type: 'reasoning',
              content: '',
              messageId: event.messageId || '',
              startTime: eventTimestamp,
            });
            break;
          }

          case 'REASONING_MESSAGE_CONTENT': {
            const activity = getCurrentActivity();
            if (activity) {
              const lastFragment =
                activity.fragments[activity.fragments.length - 1];
              if (lastFragment && lastFragment.type === 'reasoning') {
                lastFragment.content += event.delta || '';
              }
            } else {
              const lastMsg = messages[messages.length - 1];
              if (lastMsg && lastMsg.type === 'reasoning') {
                lastMsg.content += event.delta || '';
              }
            }
            break;
          }

          case 'REASONING_MESSAGE_END': {
            const activity = getCurrentActivity();
            if (activity) {
              const lastFragment =
                activity.fragments[activity.fragments.length - 1];
              if (
                lastFragment &&
                lastFragment.type === 'reasoning' &&
                lastFragment.startTime
              ) {
                lastFragment.duration = formatDuration(
                  eventTimestamp - lastFragment.startTime,
                );
              }
            } else {
              const lastMsg = messages[messages.length - 1];
              if (
                lastMsg &&
                lastMsg.type === 'reasoning' &&
                lastMsg.startTime
              ) {
                lastMsg.duration = formatDuration(
                  eventTimestamp - lastMsg.startTime,
                );
              }
            }
            break;
          }

          case 'TEXT_MESSAGE_START': {
            addFragment({
              type: 'text',
              content: '',
              messageId: event.messageId || '',
            });
            break;
          }

          case 'TEXT_MESSAGE_CONTENT': {
            const activity = getCurrentActivity();
            if (activity) {
              const lastFragment =
                activity.fragments[activity.fragments.length - 1];
              if (lastFragment && lastFragment.type === 'text') {
                lastFragment.content += event.delta || '';
              }
            } else {
              const lastMsg = messages[messages.length - 1];
              if (lastMsg && lastMsg.type === 'text') {
                lastMsg.content += event.delta || '';
              }
            }
            break;
          }

          case 'TEXT_MESSAGE_END':
            break;

          case 'TOOL_CALL_START': {
            const toolCallData = {
              id: event.toolCallId,
              name: event.toolCallName,
              arguments: '',
              status: 'completed',
              result: '',
              startTime: eventTimestamp,
              executionTime: '',
            };
            toolCallMap.set(event.toolCallId, toolCallData);
            addFragment({
              type: 'tool_call',
              toolCall: toolCallData,
              messageId: event.messageId || '',
            });
            break;
          }

          case 'TOOL_CALL_ARGS': {
            if (toolCallMap.has(event.toolCallId)) {
              const toolCall = toolCallMap.get(event.toolCallId);
              toolCall.arguments += event.delta || '';
            }
            break;
          }

          case 'TOOL_CALL_END': {
            // 不删除 toolCallMap，等 TOOL_CALL_RESULT 处理
            break;
          }

          case 'TOOL_CALL_RESULT': {
            let executionTime = '';
            if (toolCallMap.has(event.toolCallId)) {
              const toolCall = toolCallMap.get(event.toolCallId);
              toolCall.result = event.content || '';
              toolCall.status = 'completed';
              if (toolCall.startTime && eventTimestamp) {
                executionTime = formatDuration(
                  eventTimestamp - toolCall.startTime,
                );
                toolCall.executionTime = executionTime;
              }
              toolCallMap.delete(event.toolCallId);
            }
            const activity = getCurrentActivity();
            if (activity) {
              const fragment = activity.fragments.find(
                f =>
                  f.type === 'tool_call' && f.toolCall?.id === event.toolCallId,
              );
              if (fragment && fragment.toolCall) {
                fragment.toolCall.result = event.content || '';
                fragment.toolCall.status = 'completed';
                fragment.toolCall.executionTime = executionTime;
              }
            } else {
              const toolCallMsg = messages.find(
                m =>
                  m.type === 'tool_call' && m.toolCall?.id === event.toolCallId,
              );
              if (toolCallMsg && toolCallMsg.toolCall) {
                toolCallMsg.toolCall.result = event.content || '';
                toolCallMsg.toolCall.status = 'completed';
                toolCallMsg.toolCall.executionTime = executionTime;
              }
            }
            break;
          }
        }
      }

      // 处理未关闭的 activity
      while (activityStack.length > 0) {
        const activity = activityStack.pop();
        if (activityStack.length > 0) {
          activityStack[activityStack.length - 1].fragments.push(activity);
        } else {
          messages.push({
            id: this.generateId(),
            role: 'assistant',
            ...activity,
            timestamp: Date.now(),
          });
        }
      }

      return this.mergeToFragments(messages);
    },

    // 将消息合并为带 fragments 的格式
    mergeToFragments(messages) {
      const result = [];
      let currentAssistant = null;

      for (const msg of messages) {
        if (msg.role === 'user') {
          if (currentAssistant) {
            result.push(currentAssistant);
            currentAssistant = null;
          }
          result.push(msg);
        } else if (msg.role === 'assistant') {
          if (!currentAssistant) {
            currentAssistant = {
              id: msg.id || this.generateId(),
              role: 'assistant',
              content: '',
              reasoning: '',
              toolCalls: [],
              fragments: [],
            };
          }

          if (msg.type === 'activity') {
            currentAssistant.fragments.push({
              type: 'activity',
              activityType: msg.activityType,
              agentName: msg.agentName,
              fragments: msg.fragments || [],
            });
          } else if (msg.type === 'reasoning') {
            currentAssistant.fragments.push({
              type: 'reasoning',
              content: msg.content,
              duration: msg.duration || '',
            });
            currentAssistant.reasoning = msg.content;
          } else if (msg.type === 'tool_call' && msg.toolCall) {
            currentAssistant.fragments.push({
              type: 'tool_call',
              toolCall: msg.toolCall,
            });
            currentAssistant.toolCalls.push(msg.toolCall);
          } else if (msg.type === 'workspace') {
            currentAssistant.fragments.push({
              type: 'workspace',
              workspaceInfo: msg.workspaceInfo,
              runId: msg.runId,
            });
          } else if (msg.type === 'text' && msg.content) {
            currentAssistant.fragments.push({
              type: 'text',
              content: msg.content,
            });
            currentAssistant.content += msg.content;
          }
        }
      }

      if (currentAssistant) {
        result.push(currentAssistant);
      }

      return result;
    },

    parseDuration(durationStr) {
      if (!durationStr) return 0;
      const match = durationStr.match(/(\d+)m\s*(\d+)s/);
      if (match) {
        return parseInt(match[1]) * 60000 + parseInt(match[2]) * 1000;
      }
      const seconds = durationStr.match(/(\d+)s/);
      if (seconds) {
        return parseInt(seconds[1]) * 1000;
      }
      const ms = durationStr.match(/(\d+)ms/);
      if (ms) {
        return parseInt(ms[1]);
      }
      return 0;
    },

    formatMessage(msg) {
      if (!msg) return null;

      // 如果已经是标准格式
      if (msg.role && (msg.content || msg.toolCalls || msg.reasoning)) {
        return {
          id: msg.id || this.generateId(),
          role: msg.role,
          content: this.formatContent(msg.content),
          toolCalls: msg.toolCalls || null,
          toolResults: msg.toolResults || null,
          toolCallId: msg.toolCallId || null,
          reasoning: msg.reasoning || '',
          reasoningDuration: msg.reasoningDuration || '',
          toolDuration: msg.toolDuration || '',
        };
      }

      // 处理 AG-UI 协议格式
      if (msg.type) {
        switch (msg.type) {
          case 'TEXT_MESSAGE':
          case 'text_message':
            return {
              id: msg.id || msg.messageId || this.generateId(),
              role: msg.role || 'assistant',
              content: this.formatContent(msg.content || msg.text),
              toolCalls: null,
              toolResults: null,
              toolCallId: null,
              reasoning: '',
              reasoningDuration: '',
              toolDuration: '',
            };
          case 'TOOL_CALL':
          case 'tool_call':
            return {
              id: msg.id || this.generateId(),
              role: 'tool',
              content: this.formatContent(msg.result || msg.content),
              toolCalls: null,
              toolResults: null,
              toolCallId: msg.toolCallId || msg.id,
              reasoning: '',
              reasoningDuration: '',
              toolDuration: '',
            };
          default:
            // 尝试从 message 字段获取内容
            if (msg.message) {
              return this.formatMessage(msg.message);
            }
            return null;
        }
      }

      // 尝试处理嵌套结构
      if (msg.message) {
        return this.formatMessage(msg.message);
      }

      // 跳过无效消息
      if (!msg.role && !msg.content && !msg.text) {
        return null;
      }

      return {
        id: msg.id || this.generateId(),
        role: msg.role || 'unknown',
        content: this.formatContent(msg.content || msg.text || ''),
        toolCalls: msg.toolCalls || null,
        toolResults: msg.toolResults || null,
        toolCallId: msg.toolCallId || null,
        reasoning: msg.reasoning || '',
        reasoningDuration: msg.reasoningDuration || '',
        toolDuration: msg.toolDuration || '',
      };
    },

    async loadConfig() {
      if (!this.currentThreadId) return;
      try {
        const res = await getGeneralAgentConfig({
          threadId: this.currentThreadId,
        });
        if (res.code === 0 && res.data) {
          if (res.data.modelConfig) {
            const modelConfig = res.data.modelConfig;
            this.selectedModel = modelConfig.modelId || modelConfig.model || '';
          }
          if (res.data.toolList && Array.isArray(res.data.toolList)) {
            this.selectedTools = res.data.toolList.map(tool => ({
              toolId: tool.toolId,
              toolName: tool.toolName,
              toolType: tool.toolType,
              enable: tool.enable,
            }));
          }
          if (res.data.assistantList && Array.isArray(res.data.assistantList)) {
            this.selectedAssistants = res.data.assistantList.map(assistant => ({
              assistantId: assistant.agentId || assistant.assistantId,
              name: assistant.name,
            }));
          }
        }
      } catch (error) {
        console.error('加载配置失败:', error);
      }
    },

    async saveModelConfig(silent = false) {
      if (!this.currentThreadId) {
        return;
      }
      if (!this.selectedModel) {
        if (!silent) {
          this.$message.warning('请选择模型');
        }
        return;
      }
      try {
        const selectedModelConfig = this.modelList.find(
          m => m.modelId === this.selectedModel,
        );
        const res = await updateGeneralAgentConfig({
          threadId: this.currentThreadId,
          modelConfig: {
            modelId: this.selectedModel,
            model: selectedModelConfig?.model || '',
            provider: selectedModelConfig?.provider || '',
            displayName: selectedModelConfig?.modelName || '',
            modelType: selectedModelConfig?.modelType || 'llm',
            config: selectedModelConfig?.config || {},
          },
          toolList: this.selectedTools.map(t => ({
            toolId: t.toolId,
            toolType: t.toolType,
          })),
        });
        if (res.code === 0) {
          if (!silent) {
            this.$message.success('配置已保存');
          }
        } else {
          this.$message.error(res.msg || '保存模型配置失败');
        }
      } catch (error) {
        console.error('保存模型配置失败:', error);
        this.$message.error('保存模型配置失败');
      }
    },

    formatContent(content) {
      if (typeof content === 'string') return content;
      if (Array.isArray(content)) {
        return content
          .filter(item => item.type === 'text')
          .map(item => item.text)
          .join('\n');
      }
      if (typeof content === 'object' && content?.text) return content.text;
      return '';
    },

    extractFilesFromContent(content) {
      if (!Array.isArray(content)) return null;
      const files = content.filter(item => item.type === 'binary');
      if (files.length === 0) return null;
      return files.map(file => ({
        name: file.fileName || 'unknown',
        type: file.mimeType || 'application/octet-stream',
        url: file.url,
        displayUrl: file.url,
      }));
    },

    handleKeyDown(e) {
      if (e.shiftKey) return;
      e.preventDefault();
      this.sendMessage();
    },

    // 将内部服务地址转换为外部可访问地址
    convertToExternalUrl(url) {
      if (!url) return url;
      // 替换 minio 内部服务名为外部地址
      return url.replace(/minio-wanwu:9000/g, '192.168.0.21:9000');
    },

    async handleFileChange(file) {
      // 先显示本地预览
      const localUrl = URL.createObjectURL(file.raw);
      const tempFile = {
        name: file.name,
        type: file.raw.type,
        url: localUrl,
        localUrl: localUrl,
        uploading: true,
        uploadProgress: 0,
      };
      this.uploadedFiles.push(tempFile);

      // 上传文件到服务器
      try {
        const res = await uploadGeneralAgentFile(file.raw, percent => {
          // 更新进度
          const index = this.uploadedFiles.findIndex(
            f => f.localUrl === localUrl,
          );
          if (index !== -1) {
            this.$set(this.uploadedFiles[index], 'uploadProgress', percent);
          }
        });
        if (res.code === 0 && res.data?.files?.[0]?.filePath) {
          // 更新为服务器 URL（转换为外部可访问地址）
          const index = this.uploadedFiles.findIndex(
            f => f.localUrl === localUrl,
          );
          if (index !== -1) {
            // 使用 Vue.set 确保响应式更新
            this.$set(this.uploadedFiles, index, {
              name: file.name,
              type: file.raw.type,
              url: res.data.files[0].filePath, // 原始 minio URL，用于发送给后端
              fileName: res.data.files[0].fileName, // 服务器返回的文件名
              displayUrl: this.convertToExternalUrl(res.data.files[0].filePath), // 转换后的 URL，用于前端显示
              uploading: false,
              uploadProgress: 100,
            });
          }
          URL.revokeObjectURL(localUrl);
        } else {
          // 上传失败，移除文件
          const index = this.uploadedFiles.findIndex(
            f => f.localUrl === localUrl,
          );
          if (index !== -1) {
            this.uploadedFiles.splice(index, 1);
          }
          this.$message.error(res.msg || '文件上传失败');
          URL.revokeObjectURL(localUrl);
        }
      } catch (error) {
        console.error('文件上传失败:', error);
        const index = this.uploadedFiles.findIndex(
          f => f.localUrl === localUrl,
        );
        if (index !== -1) {
          this.uploadedFiles.splice(index, 1);
        }
        this.$message.error('文件上传失败');
        URL.revokeObjectURL(localUrl);
      }
    },

    removeFile(index) {
      this.uploadedFiles.splice(index, 1);
    },

    handleModelChange() {
      this.saveModelConfig();
    },

    isToolSelected(toolId) {
      return this.selectedTools.some(t => t.toolId === toolId);
    },

    async toggleTool(tool) {
      const index = this.selectedTools.findIndex(t => t.toolId === tool.toolId);
      if (index > -1) {
        this.selectedTools.splice(index, 1);
      } else {
        this.selectedTools.push({
          toolId: tool.toolId,
          toolName: tool.toolName,
          toolType: tool.toolType,
        });
      }
      // 静默保存配置，不显示消息
      await this.saveModelConfig(true);
    },

    async sendMessage() {
      const content = this.inputMessage.trim();
      if (!content && this.uploadedFiles.length === 0) return;

      // 检查当前会话是否正在流式传输
      const currentStreaming = this.streamingMap[this.currentThreadId];
      if (currentStreaming && currentStreaming.isStreaming) return;

      // 检查是否有文件正在上传
      const uploadingFiles = this.uploadedFiles.filter(f => f.uploading);
      if (uploadingFiles.length > 0) {
        this.$message.warning('请等待文件上传完成');
        return;
      }

      if (this.isNewConversation || !this.currentThreadId) {
        const title = content.slice(0, 50);
        const threadId = await this.createConversationWithTitle(title);
        if (!threadId) {
          this.$message.error('创建对话失败，请重试');
          return;
        }
      }

      const userMessage = this.buildUserMessage(content);

      // 确保当前会话的消息列表存在
      if (!this.messagesMap[this.currentThreadId]) {
        this.$set(this.messagesMap, this.currentThreadId, []);
      }

      // 添加用户消息到当前会话
      const messages = this.messagesMap[this.currentThreadId];
      messages.push({
        id: this.generateId(),
        role: 'user',
        content: content,
        files: [...this.uploadedFiles],
      });

      this.inputMessage = '';
      this.uploadedFiles = [];
      this.$nextTick(() => this.scrollToBottom());

      await this.startStreaming(userMessage);
    },

    buildUserMessage(content) {
      const message = { id: this.generateId(), role: 'user' };

      // 如果没有文件，直接返回文本
      if (this.uploadedFiles.length === 0) {
        message.content = content;
        return message;
      }

      // 有文件时，构建多部分内容
      const contentArray = [];

      // 添加文本内容（如果有）
      if (content && content.trim()) {
        contentArray.push({ type: 'text', text: content.trim() });
      }

      // 添加文件内容 - 后端统一使用 type: 'binary'，根据 mimeType 判断具体类型
      this.uploadedFiles.forEach(file => {
        contentArray.push({
          type: 'binary',
          mimeType: file.type || 'application/octet-stream',
          url: file.url, // 使用服务器返回的 HTTP URL
          fileName: file.fileName, // 服务器返回的文件名
        });
      });

      message.content = contentArray;
      return message;
    },

    async startStreaming(userMessage) {
      if (!this.currentThreadId) {
        this.$message.error('对话ID不存在，请刷新页面重试');
        return;
      }

      const streamingThreadId = this.currentThreadId;

      // 初始化该会话的流式状态
      const abortController = new AbortController();
      const assistantMessage = {
        id: this.generateId(),
        role: 'assistant',
        content: '',
        reasoning: '',
        toolCalls: [],
        toolResults: [],
        fragments: [],
        isStreaming: true,
        threadId: streamingThreadId,
      };

      // 设置该会话的流式状态
      this.$set(this.streamingMap, streamingThreadId, {
        isStreaming: true,
        abortController: abortController,
        streamingMessage: assistantMessage,
        activityStack: [],
        currentActivity: null,
        currentFragment: null,
        toolCallMap: new Map(),
      });

      // 确保该会话的消息列表存在
      if (!this.messagesMap[streamingThreadId]) {
        this.$set(this.messagesMap, streamingThreadId, []);
      }

      // 添加消息到对应会话的消息列表
      const messages = this.messagesMap[streamingThreadId];
      messages.push(assistantMessage);

      this.currentStage = 'understanding';

      // 重置滚动状态
      this.userHasScrolled = false;
      this.showScrollToBottom = false;

      const parser = new SSEEventParser();

      try {
        await chatGeneralAgentConversation({
          threadId: streamingThreadId,
          messages: [userMessage],
          onMessage: event => {
            // 直接更新对应会话的消息，不检查当前会话
            this.handleSSEEvent(
              event,
              assistantMessage,
              parser,
              streamingThreadId,
            );
          },
          onError: error => {
            console.error('SSE Error:', error);
            // 只在对应会话显示错误提示
            if (this.currentThreadId === streamingThreadId) {
              this.$message.error('对话请求失败');
            }
            // 更新该会话的流式状态
            const streaming = this.streamingMap[streamingThreadId];
            if (streaming) {
              streaming.isStreaming = false;
              streaming.streamingMessage = null;
            }
            assistantMessage.isStreaming = false;
          },
          signal: abortController.signal,
        });
      } catch (error) {
        console.error('Stream error:', error);
        if (
          error.name !== 'AbortError' &&
          this.currentThreadId === streamingThreadId
        ) {
          this.$message.error('发送消息失败: ' + (error.message || error));
        }
      } finally {
        // 更新该会话的流式状态
        const streaming = this.streamingMap[streamingThreadId];
        if (streaming) {
          streaming.isStreaming = false;
          streaming.streamingMessage = null;
          streaming.abortController = null;
        }
        assistantMessage.isStreaming = false;
        this.currentStage = '';
        // 流式结束后滚动到底部
        if (this.currentThreadId === streamingThreadId) {
          this.userHasScrolled = false;
          this.showScrollToBottom = false;
          this.$nextTick(() => this.scrollToBottom(true));
        }
      }
    },

    handleSSEEvent(event, assistantMessage, parser, streamingThreadId) {
      const parsed = parser.parse(event);
      if (!parsed) return;

      const streamState = this.streamingMap[streamingThreadId];
      if (!streamState) return;

      const getCurrentFragments = () => {
        if (streamState.currentActivity) {
          return streamState.currentActivity.fragments;
        }
        return assistantMessage.fragments;
      };

      const addFragment = fragment => {
        const fragments = getCurrentFragments();
        if (fragments) {
          fragments.push(fragment);
        }
      };

      switch (parsed.type) {
        case 'RUN_STARTED':
          this.currentRunId = parsed.runId;
          if (this.currentThreadId === streamingThreadId) {
            this.currentStage = 'understanding';
          }
          break;

        case 'ACTIVITY_SNAPSHOT': {
          const activityContent = parsed.content || {};
          if (parsed.activityType === ActivityType.SUB_AGENT) {
            if (activityContent.status === ActivityStatus.STARTED) {
              const activity = {
                type: 'activity',
                activityType: parsed.activityType,
                activityId: parsed.activityId || '',
                agentName: activityContent.agentName || '',
                fragments: [],
                isStreaming: true,
                startTime: Date.now(),
                duration: '',
              };
              // 先添加到父级 fragments，再设置 currentActivity
              const parentFragments = getCurrentFragments();
              if (parentFragments) {
                parentFragments.push(activity);
              }
              streamState.currentActivity = activity;
              streamState.activityStack.push(activity);
            } else if (activityContent.status === ActivityStatus.FINISHED) {
              if (streamState.activityStack.length > 0) {
                const finishedActivity = streamState.activityStack.pop();
                finishedActivity.isStreaming = false;
                if (finishedActivity.startTime) {
                  finishedActivity.duration = formatDuration(
                    Date.now() - finishedActivity.startTime,
                  );
                }
                if (streamState.activityStack.length > 0) {
                  streamState.currentActivity =
                    streamState.activityStack[
                      streamState.activityStack.length - 1
                    ];
                } else {
                  streamState.currentActivity = null;
                }
              }
            }
          } else if (
            parsed.activityType === ActivityType.WORKSPACE &&
            activityContent.runId
          ) {
            this.handleWorkspaceActivity({
              runId: activityContent.runId,
              threadId: activityContent.threadId || this.currentThreadId,
              fileCount: activityContent.fileCount || 0,
              totalSize: activityContent.totalSize || 0,
              timestamp: activityContent.timestamp || Date.now(),
            });
            addFragment({
              type: 'workspace',
              workspaceInfo: {
                fileCount: activityContent.fileCount || 0,
                totalSize: activityContent.totalSize || 0,
              },
              runId: activityContent.runId,
            });
            if (this.currentThreadId === streamingThreadId) {
              this.$notify({
                type: 'success',
                title: '工作空间已更新',
                message: `生成了 ${activityContent.fileCount || 0} 个文件`,
                duration: 3000,
                onClick: () => {
                  this.showPanel();
                },
              });
            }
          }
          break;
        }

        case 'REASONING_MESSAGE_START':
          streamState.currentFragment = {
            type: 'reasoning',
            content: '',
            messageId: parsed.messageId || '',
            startTime: Date.now(),
            isStreaming: true,
          };
          addFragment(streamState.currentFragment);
          if (this.currentThreadId === streamingThreadId) {
            this.currentStage = 'thinking';
          }
          break;

        case 'REASONING_MESSAGE_CONTENT':
          if (
            streamState.currentFragment &&
            streamState.currentFragment.type === 'reasoning'
          ) {
            streamState.currentFragment.content += parsed.delta || '';
            if (!streamState.currentActivity) {
              assistantMessage.reasoning += parsed.delta || '';
            }
          }
          break;

        case 'REASONING_MESSAGE_END':
          if (
            streamState.currentFragment &&
            streamState.currentFragment.type === 'reasoning'
          ) {
            if (streamState.currentFragment.startTime) {
              streamState.currentFragment.duration = formatDuration(
                Date.now() - streamState.currentFragment.startTime,
              );
            }
            streamState.currentFragment.isStreaming = false;
          }
          streamState.currentFragment = null;
          break;

        case 'TEXT_MESSAGE_START':
          streamState.currentFragment = {
            type: 'text',
            content: '',
            messageId: parsed.messageId || '',
            isStreaming: true,
          };
          addFragment(streamState.currentFragment);
          assistantMessage.id = parsed.messageId || assistantMessage.id;
          if (this.currentThreadId === streamingThreadId) {
            this.currentStage = 'generating';
          }
          break;

        case 'TEXT_MESSAGE_CONTENT':
          if (
            streamState.currentFragment &&
            streamState.currentFragment.type === 'text'
          ) {
            streamState.currentFragment.content += parsed.delta || '';
            if (!streamState.currentActivity) {
              assistantMessage.content += parsed.delta || '';
            }
          }
          break;

        case 'TEXT_MESSAGE_END':
          if (
            streamState.currentFragment &&
            streamState.currentFragment.type === 'text'
          ) {
            streamState.currentFragment.isStreaming = false;
          }
          streamState.currentFragment = null;
          break;

        case 'TOOL_CALL_START': {
          const toolCallData = {
            id: parsed.toolCallId,
            name: parsed.toolCallName,
            arguments: '',
            status: 'running',
            result: '',
            startTime: Date.now(),
            executionTime: '',
          };
          streamState.toolCallMap.set(parsed.toolCallId, toolCallData);
          assistantMessage.toolCalls.push(toolCallData);
          streamState.currentFragment = {
            type: 'tool_call',
            toolCall: toolCallData,
            messageId: parsed.messageId || '',
          };
          addFragment(streamState.currentFragment);
          if (this.currentThreadId === streamingThreadId) {
            this.currentStage = 'tool_calling';
          }
          break;
        }

        case 'TOOL_CALL_ARGS':
          if (streamState.toolCallMap.has(parsed.toolCallId)) {
            const toolCall = streamState.toolCallMap.get(parsed.toolCallId);
            toolCall.arguments += parsed.delta || '';
          }
          break;

        case 'TOOL_CALL_END':
          // 不设置 completed，不计算时间，等 TOOL_CALL_RESULT
          streamState.currentFragment = null;
          break;

        case 'TOOL_CALL_RESULT':
          if (streamState.toolCallMap.has(parsed.toolCallId)) {
            const toolCall = streamState.toolCallMap.get(parsed.toolCallId);
            toolCall.result = parsed.content || '';
            toolCall.status = 'completed';
            if (toolCall.startTime) {
              toolCall.executionTime = formatDuration(
                Date.now() - toolCall.startTime,
              );
            }
            streamState.toolCallMap.delete(parsed.toolCallId);
          }
          const tc = assistantMessage.toolCalls.find(
            t => t.id === parsed.toolCallId,
          );
          if (tc) {
            tc.result = parsed.content || '';
            tc.status = 'completed';
            if (tc.startTime) {
              tc.executionTime = formatDuration(Date.now() - tc.startTime);
            }
          }
          const fragments = getCurrentFragments();
          const toolCallFragment = fragments.find(
            f => f.type === 'tool_call' && f.toolCall?.id === parsed.toolCallId,
          );
          if (toolCallFragment && toolCallFragment.toolCall) {
            toolCallFragment.toolCall.result = parsed.content || '';
            toolCallFragment.toolCall.status = 'completed';
            if (toolCallFragment.toolCall.startTime) {
              toolCallFragment.toolCall.executionTime = formatDuration(
                Date.now() - toolCallFragment.toolCall.startTime,
              );
            }
          }
          break;

        case 'RUN_FINISHED':
          while (streamState.activityStack.length > 0) {
            const activity = streamState.activityStack.pop();
            if (streamState.activityStack.length > 0) {
              streamState.activityStack[
                streamState.activityStack.length - 1
              ].fragments.push(activity);
            } else {
              assistantMessage.fragments.push(activity);
            }
          }
          streamState.currentActivity = null;
          streamState.currentFragment = null;
          break;
      }

      if (this.currentThreadId === streamingThreadId) {
        this.$nextTick(() => this.scrollToBottom());
      }
    },

    async loadWorkspaceFiles() {
      if (!this.activeWorkspace || !this.currentThreadId) return;

      this.workspaceLoading = true;
      try {
        const res = await getGeneralAgentWorkspace({
          threadId: this.currentThreadId,
          runId: this.activeWorkspace.runId,
        });
        if (res.code === 0 && res.data) {
          this.setWorkspaceTree({
            threadId: this.currentThreadId,
            runId: this.activeWorkspace.runId,
            data: res.data,
          });
        }
      } catch (error) {
        console.error('加载工作空间文件失败:', error);
      } finally {
        this.workspaceLoading = false;
      }
    },

    toggleWorkspacePanel() {
      if (this.panelVisible) {
        this.hidePanel();
      } else {
        this.showPanel();
        if (this.activeWorkspace) {
          this.loadWorkspaceFiles();
        }
      }
    },

    handleViewWorkspace(data) {
      this.setActiveWorkspace({
        runId: data.runId,
        threadId: data.threadId || this.currentThreadId,
        fileCount: data.fileCount || 0,
        totalSize: data.totalSize || 0,
        timestamp: Date.now(),
      });
      this.showPanel();
    },

    stopStreaming() {
      // 中止当前会话的 SSE 流
      const streaming = this.streamingMap[this.currentThreadId];
      if (streaming && streaming.abortController) {
        streaming.abortController.abort();
        streaming.isStreaming = false;
        streaming.streamingMessage = null;
        streaming.abortController = null;
      }
    },

    scrollToBottom(force = false) {
      if (!force && this.userHasScrolled) {
        this.showScrollToBottom = true;
        return;
      }
      this.isAutoScrolling = true;
      const container = this.$refs.messageArea;
      if (container) {
        container.scrollTop = container.scrollHeight;
      }
      setTimeout(() => {
        this.isAutoScrolling = false;
      }, 100);
    },

    handleMessageAreaScroll() {
      if (this.isAutoScrolling) return;

      const container = this.$refs.messageArea;
      if (!container) return;

      const { scrollTop, scrollHeight, clientHeight } = container;
      const distanceFromBottom = scrollHeight - scrollTop - clientHeight;
      const threshold = 150;

      const isNearBottom = distanceFromBottom < threshold;

      if (isNearBottom) {
        this.userHasScrolled = false;
        this.showScrollToBottom = false;
      } else {
        this.userHasScrolled = true;
        this.showScrollToBottom = true;
      }
    },

    handleScrollToBottomClick() {
      this.userHasScrolled = false;
      this.showScrollToBottom = false;
      this.scrollToBottom(true);
    },

    generateId() {
      return (
        'msg_' + Date.now() + '_' + Math.random().toString(36).substr(2, 9)
      );
    },

    getToolResultsForMessage(message) {
      if (message.toolResults && message.toolResults.length > 0) {
        return message.toolResults;
      }
      return [];
    },

    getPreviewType(file) {
      if (!file || !file.name) return 'unsupported';
      const ext = file.name.split('.').pop().toLowerCase();

      const typeMap = {
        image: ['png', 'jpg', 'jpeg', 'gif', 'svg', 'webp', 'bmp', 'ico'],
        video: ['mp4', 'webm', 'ogg', 'mov', 'm4v', 'avi', 'mkv'],
        audio: ['mp3', 'wav', 'ogg', 'm4a', 'flac', 'aac', 'wma'],
        pdf: ['pdf'],
        ppt: ['ppt', 'pptx'],
        excel: ['xls', 'xlsx'],
        office: ['doc', 'docx'],
        html: ['html', 'htm'],
        markdown: ['md'],
        text: [
          'txt',
          'json',
          'js',
          'ts',
          'jsx',
          'tsx',
          'vue',
          'py',
          'java',
          'go',
          'rs',
          'c',
          'cpp',
          'h',
          'hpp',
          'cs',
          'rb',
          'php',
          'swift',
          'kt',
          'scala',
          'css',
          'scss',
          'sass',
          'less',
          'xml',
          'yaml',
          'yml',
          'toml',
          'ini',
          'conf',
          'cfg',
          'sh',
          'bash',
          'zsh',
          'bat',
          'sql',
          'dockerfile',
          'makefile',
          'r',
          'm',
          'lua',
          'pl',
          'pm',
        ],
      };

      for (const [type, exts] of Object.entries(typeMap)) {
        if (exts.includes(ext)) {
          return type;
        }
      }

      return 'unsupported';
    },

    async handlePreviewFile(data) {
      const { file, filePath, threadId, runId } = data;

      this.previewFile = file;
      this.previewLoading = true;
      this.previewVisible = true;
      this.previewUrl = '';
      this.previewContent = '';
      this.previewType = '';
      this.previewBlobUrl = '';
      this.previewExcelData = null;

      try {
        this.previewFilePath = filePath;
        this.previewFileExt = file.name.split('.').pop().toLowerCase();
        const blob = await previewGeneralAgentWorkspace({
          threadId,
          runId,
          path: filePath,
        });

        this.previewType = this.getPreviewType(file);

        if (
          ['image', 'video', 'audio', 'pdf', 'html'].includes(this.previewType)
        ) {
          this.previewBlobUrl = URL.createObjectURL(blob);
          this.previewUrl = this.previewBlobUrl;
        } else if (this.previewType === 'ppt') {
          this.previewUrl = blob;
          this.previewBlobUrl = blob;
        } else if (['markdown', 'text'].includes(this.previewType)) {
          this.previewContent = await blob.text();
        } else if (this.previewType === 'excel') {
          const arrayBuffer = await blob.arrayBuffer();
          const workbook = XLSX.read(arrayBuffer, { type: 'array' });
          const excelData = workbook.SheetNames.map(sheetName => {
            const sheet = workbook.Sheets[sheetName];
            const jsonData = XLSX.utils.sheet_to_json(sheet, {
              header: 1,
              defval: '',
            });
            const merges = sheet['!merges'] || [];
            return {
              name: sheetName,
              data: jsonData,
              merges: merges.map(m => ({
                sr: m.s.r,
                sc: m.s.c,
                er: m.e.r,
                ec: m.e.c,
              })),
              colCount: Math.max(
                1,
                sheet['!ref']
                  ? XLSX.utils.decode_range(sheet['!ref']).e.c + 1
                  : 1,
              ),
            };
          });
          this.previewExcelData = excelData;
        }
      } catch (error) {
        console.error('预览文件失败:', error);
        this.$message.error('预览文件失败');
        this.previewType = 'unsupported';
      } finally {
        this.previewLoading = false;
      }
    },

    closePreview() {
      if (this.previewBlobUrl) {
        URL.revokeObjectURL(this.previewBlobUrl);
        this.previewBlobUrl = '';
      }
    },

    async downloadPreviewFile(file) {
      if (!file || !this.previewFilePath) return;

      try {
        const blob = await downloadGeneralAgentWorkspace({
          threadId: this.activeWorkspace?.threadId || this.currentThreadId,
          runId: this.activeWorkspace?.runId || this.currentRunId,
          path: this.previewFilePath,
        });
        const url = URL.createObjectURL(blob);
        const a = document.createElement('a');
        a.href = url;
        a.download = file.name;
        document.body.appendChild(a);
        a.click();
        document.body.removeChild(a);
        URL.revokeObjectURL(url);
        this.$message.success('下载成功');
      } catch (error) {
        console.error('下载文件失败:', error);
        this.$message.error('下载文件失败');
      }
    },

    // 重新生成 - 找到上一条用户消息并重新发送
    handleRegenerate(message) {
      if (this.isStreaming) return;

      // 找到这条助手消息的索引
      const messageIndex = this.messageList.findIndex(m => m.id === message.id);
      if (messageIndex <= 0) return;

      // 找到上一条用户消息
      let userMessage = null;
      for (let i = messageIndex - 1; i >= 0; i--) {
        if (this.messageList[i].role === 'user') {
          userMessage = this.messageList[i];
          break;
        }
      }

      if (!userMessage) return;

      // 删除当前助手消息（保留用户消息和之前的消息）
      this.messageList.splice(messageIndex, 1);

      // 构建请求消息
      const requestMessage = this.buildRequestMessage(userMessage);

      // 直接调用 startStreaming，不再添加用户消息
      this.$nextTick(() => {
        this.startStreaming(requestMessage);
      });
    },

    // 根据已存在的用户消息构建请求消息
    buildRequestMessage(userMessage) {
      const message = { id: this.generateId(), role: 'user' };

      // 如果没有文件，直接返回文本
      if (!userMessage.files || userMessage.files.length === 0) {
        message.content = userMessage.content;
        return message;
      }

      // 有文件时，构建多部分内容
      const contentArray = [];

      // 添加文本内容（如果有）
      if (userMessage.content && userMessage.content.trim()) {
        contentArray.push({ type: 'text', text: userMessage.content.trim() });
      }

      // 添加文件内容
      userMessage.files.forEach(file => {
        contentArray.push({
          type: 'binary',
          mimeType: file.type || 'application/octet-stream',
          url: file.url,
          fileName: file.name,
        });
      });

      message.content = contentArray;
      return message;
    },

    async handleCommand(command, item) {
      if (command === 'delete') {
        try {
          await this.$confirm('确定要删除这个对话吗？', '提示', {
            type: 'warning',
          });
          const res = await deleteGeneralAgentConversation({
            threadId: item.threadId,
          });
          if (res.code === 0) {
            this.$message.success('删除成功');
            if (this.currentThreadId === item.threadId) {
              this.currentThreadId = '';
              this.isNewConversation = true;
              this.messageList = [];
              this.hidePanel();
            }
            this.fetchConversationList();
          }
        } catch (error) {
          if (error !== 'cancel') console.error('删除对话失败:', error);
        }
      }
    },
  },
};
</script>

<style lang="scss" scoped>
$claude-primary: #10a37f;
$claude-primary-light: #1ab38b;
$claude-primary-dark: #0d8a6a;
$claude-bg: #ffffff;
$claude-bg-secondary: #f7f7f8;
$claude-border: #e5e5e5;
$claude-text: #1a1a1a;
$claude-text-secondary: #666666;
$claude-text-muted: #999999;
$message-max-width: 900px;

.general-agent-page {
  display: flex;
  position: absolute;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background: #f5f7fa;
  overflow: hidden;
  padding: 16px;
  box-sizing: border-box;
}

.sidebar {
  display: flex;
  flex-direction: column;
  flex-shrink: 0;
  width: 240px;
  height: 100%;
  background: #fff;
  border-radius: 12px;
  overflow: hidden;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.04);
  transition:
    width 0.3s ease,
    margin-right 0.3s ease;
  margin-right: 16px;

  &.collapsed {
    width: 0;
    margin-right: 0;
    box-shadow: none;
  }

  .sidebar-content {
    display: flex;
    flex-direction: column;
    width: 240px;
    height: 100%;
    flex-shrink: 0;
  }

  .sidebar-header {
    flex-shrink: 0;
    display: flex;
    align-items: center;
    justify-content: center;
    padding: 16px;
    border-bottom: 1px solid #f0f0f0;

    .new-chat-btn {
      width: 100%;
      border-radius: 12px;
      background: $claude-primary;
      border-color: $claude-primary;
      font-weight: 500;

      &:hover {
        background: $claude-primary-dark;
        border-color: $claude-primary-dark;
      }
    }
  }

  .sidebar-divider {
    height: 1px;
    background: #f0f0f0;
    flex-shrink: 0;
  }

  .conversation-list {
    flex: 1;
    overflow-y: auto;
    padding: 8px;
    min-height: 0;

    &::-webkit-scrollbar {
      width: 4px;
    }

    &::-webkit-scrollbar-track {
      background: transparent;
    }

    &::-webkit-scrollbar-thumb {
      background: #d1d5db;
      border-radius: 2px;
    }
  }

  .conversation-item {
    display: flex;
    align-items: center;
    padding: 12px 14px;
    border-radius: 10px;
    cursor: pointer;
    margin-bottom: 4px;
    transition: background-color 0.2s;

    &:hover {
      background: rgba($claude-primary, 0.08);

      .el-icon-more {
        opacity: 1;
      }
    }

    &.active {
      background: rgba($claude-primary, 0.12);

      .conversation-title {
        font-weight: 500;
      }
    }

    i:first-child {
      margin-right: 10px;
      color: $claude-text-muted;
      font-size: 16px;
    }

    .conversation-title {
      flex: 1;
      font-size: 14px;
      color: $claude-text;
      overflow: hidden;
      text-overflow: ellipsis;
      white-space: nowrap;
    }

    .el-icon-more {
      opacity: 0;
      color: $claude-text-muted;
      padding: 4px;
      transition: opacity 0.2s;

      &:hover {
        color: $claude-primary;
      }
    }
  }
}

.agent-main-content {
  flex: 1;
  display: flex;
  min-width: 0;
  min-height: 0;
  position: relative;
  overflow: hidden;
  background: #fff;
  border-radius: 12px;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.04);

  &.has-workspace {
    .main-content-body {
      flex: 1;
      min-width: 0;
    }
  }
}

.main-content-body {
  flex: 1;
  display: flex;
  flex-direction: column;
  min-width: 0;
  min-height: 0;
  position: relative;
  overflow: hidden;
}

.header {
  flex: none;
  height: 56px;
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 0 24px;
  background: #fff;
  border-bottom: 1px solid #f0f0f0;
  border-radius: 12px 12px 0 0;

  .header-left {
    display: flex;
    align-items: center;
    gap: 12px;
  }

  .header-right {
    display: flex;
    align-items: center;
    gap: 12px;
  }

  .header-title {
    font-size: 16px;
    font-weight: 600;
    color: $claude-text;
  }

  .sidebar-toggle-btn {
    display: flex;
    align-items: center;
    justify-content: center;
    width: 32px;
    height: 32px;
    border: 1px solid $claude-border;
    background: #fff;
    border-radius: 8px;
    cursor: pointer;
    color: $claude-text-muted;
    transition: all 0.2s;

    &:hover {
      border-color: $claude-primary;
      color: $claude-primary;
      background: rgba($claude-primary, 0.05);
    }

    i {
      font-size: 16px;
    }
  }
}

.message-area {
  flex: 1;
  min-height: 0;
  overflow-y: auto;
  overflow-x: hidden;
  background: #fff;
  position: relative;

  &.empty {
    flex: none;
    min-height: 0;
    height: 0;
    overflow: hidden;
  }

  .message-list {
    max-width: $message-max-width;
    margin: 0 auto;
    padding: 24px;
    min-height: 100%;
  }

  .history-loading {
    position: absolute;
    top: 0;
    left: 0;
    right: 0;
    bottom: 0;
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    color: #909399;
    font-size: 14px;
    gap: 12px;
    background: #fff;
    z-index: 10;

    i {
      font-size: 32px;
      color: #10a37f;
    }
  }
}

.scroll-to-bottom-btn {
  position: absolute;
  bottom: 120px;
  left: 50%;
  transform: translateX(-50%);
  width: 36px;
  height: 36px;
  display: flex;
  align-items: center;
  justify-content: center;
  background: #fff;
  color: #10a37f;
  border: 1px solid #10a37f;
  border-radius: 50%;
  cursor: pointer;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.15);
  transition: all 0.2s ease;
  z-index: 100;

  &:hover {
    background: #10a37f;
    color: #fff;
    transform: translateX(-50%) translateY(-2px);
    box-shadow: 0 4px 12px rgba(16, 163, 127, 0.4);
  }

  svg {
    width: 16px;
    height: 16px;
  }
}

.scroll-btn-fade-enter-active,
.scroll-btn-fade-leave-active {
  transition: all 0.3s ease;
}

.scroll-btn-fade-enter,
.scroll-btn-fade-leave-to {
  opacity: 0;
  transform: translateX(-50%) translateY(20px);
}

.input-area {
  flex: none;
  background: #fff;
  padding: 16px 24px 24px;
  border-radius: 0 0 12px 12px;

  &.is-centered {
    flex: 1;
    display: flex;
    flex-direction: column;
    justify-content: center;
    align-items: center;
    border-top: none;
    padding: 0 24px;

    .input-container {
      max-width: 800px;
      width: 100%;
    }

    .welcome-section {
      display: flex;
      flex-direction: column;
      align-items: center;
      margin-bottom: 32px;

      .welcome-avatar {
        width: 72px;
        height: 72px;
        border-radius: 20px;
        display: flex;
        align-items: center;
        justify-content: center;
        margin-bottom: 20px;
        background: #fff;
        box-shadow: 0 4px 16px rgba(0, 0, 0, 0.08);
        overflow: hidden;

        img {
          width: 100%;
          height: 100%;
          border-radius: 20px;
          object-fit: cover;
        }

        i {
          font-size: 32px;
          color: #10a37f;
        }
      }

      .welcome-title {
        font-size: 28px;
        color: $claude-text;
        font-weight: 600;
      }
    }

    .input-footer {
      display: none;
    }
  }

  &:not(.is-centered) {
    border-top: none;
  }

  .input-container {
    max-width: $message-max-width;
    margin: 0 auto;
    background: #fff;
    border-radius: 16px;
    border: 1px solid #e5e7eb;
    padding: 16px;
    box-shadow: 0 4px 16px rgba(0, 0, 0, 0.06);
    transition:
      border-color 0.2s,
      box-shadow 0.2s;

    &:focus-within {
      border-color: $claude-primary;
      box-shadow: 0 4px 24px rgba(0, 0, 0, 0.1);
    }
  }

  .file-preview {
    display: flex;
    flex-wrap: wrap;
    gap: 8px;
    margin-bottom: 12px;

    .file-item {
      position: relative;
      width: 48px;
      height: 48px;

      .file-thumb {
        width: 100%;
        height: 100%;
        object-fit: cover;
        border-radius: 8px;
      }

      .file-icon {
        width: 100%;
        height: 100%;
        display: flex;
        align-items: center;
        justify-content: center;
        background: #e5e7eb;
        border-radius: 8px;
        color: $claude-text-secondary;
      }

      .file-remove {
        position: absolute;
        top: -4px;
        right: -4px;
        width: 18px;
        height: 18px;
        background: #ef4444;
        color: #fff;
        border-radius: 50%;
        display: flex;
        align-items: center;
        justify-content: center;
        cursor: pointer;
        font-size: 10px;
        transition: transform 0.2s;
        z-index: 10;

        &:hover {
          transform: scale(1.1);
        }
      }

      .upload-overlay {
        position: absolute;
        top: 0;
        left: 0;
        right: 0;
        bottom: 0;
        background: rgba(0, 0, 0, 0.5);
        border-radius: 8px;
        display: flex;
        align-items: center;
        justify-content: center;
        z-index: 5;

        .upload-progress-bar {
          width: 32px;
          height: 32px;
          position: relative;
          display: flex;
          align-items: center;
          justify-content: center;

          svg {
            position: absolute;
            top: 0;
            left: 0;
            transform: rotate(-90deg);

            circle {
              fill: none;
              stroke-width: 3;
            }

            .progress-bg {
              stroke: rgba(255, 255, 255, 0.3);
            }

            .progress-fill {
              stroke: #fff;
              stroke-linecap: round;
              transition: stroke-dashoffset 0.3s ease;
            }
          }

          .progress-text {
            color: #fff;
            font-size: 9px;
            font-weight: 600;
            z-index: 1;
          }
        }
      }

      &.is-uploading {
        .file-remove {
          display: none;
        }
      }
    }
  }

  .input-wrapper {
    ::v-deep .el-textarea {
      .el-textarea__inner {
        background: transparent;
        border: none;
        padding: 0;
        resize: none;
        font-size: 16px;
        line-height: 1.6;
        color: $claude-text;

        &::placeholder {
          color: #9ca3af;
        }
      }
    }
  }

  .input-toolbar {
    display: flex;
    align-items: center;
    justify-content: space-between;
    margin-top: 12px;
    padding-top: 12px;
    border-top: 1px solid #f3f4f6;

    .toolbar-left {
      display: flex;
      align-items: center;
      gap: 8px;

      .model-select-inline {
        min-width: 200px;

        ::v-deep .el-input__inner {
          background: transparent;
          border: none;
          padding-left: 32px;
          font-size: 13px;
          color: $claude-text;
        }

        ::v-deep .el-input__prefix {
          left: 8px;
        }
      }
    }

    .toolbar-right {
      display: flex;
      align-items: center;
      gap: 8px;

      .action-icon {
        font-size: 18px;
        color: $claude-text-muted;
        cursor: pointer;
        padding: 8px;
        border-radius: 8px;
        transition: all 0.2s;

        &:hover {
          color: $claude-primary;
          background: rgba($claude-primary, 0.08);
        }
      }

      .send-btn {
        padding: 10px;
        border: none;
        color: #5147ff;
        line-height: 0;
        display: inline-flex;
        justify-content: center;
        align-items: center;

        &:hover {
          background-color: rgba(87, 104, 161, 0.08);
        }

        .send-icon {
          width: 18px;
          height: 18px;
          fill: currentColor;
        }

        .stop-icon {
          width: 16px;
          height: 16px;
          fill: currentColor;
        }
      }

      .stop-btn {
        background: #fff !important;
        border: 1px solid #333 !important;
        color: #333 !important;

        &:hover,
        &:focus {
          background: #333 !important;
          border-color: #333 !important;
          color: #fff !important;
        }
      }
    }
  }

  .config-btn {
    display: flex;
    align-items: center;
    gap: 4px;
    padding: 6px 12px;
    border-radius: 8px;
    cursor: pointer;
    font-size: 13px;
    color: $claude-text-secondary;
    background: transparent;
    border: 1px solid $claude-border;
    transition: all 0.2s;

    &:hover {
      background: rgba($claude-primary, 0.08);
      color: $claude-primary;
      border-color: rgba($claude-primary, 0.3);
    }

    &.has-selection {
      color: $claude-primary;
      border-color: rgba($claude-primary, 0.3);
      background: rgba($claude-primary, 0.05);
    }

    i {
      font-size: 16px;
    }

    .el-badge {
      margin-left: 4px;
    }
  }

  .model-option {
    display: flex;
    align-items: center;
    justify-content: space-between;
    width: 100%;

    .model-name {
      flex: 1;
      overflow: hidden;
      text-overflow: ellipsis;
      white-space: nowrap;
    }

    .model-provider {
      flex-shrink: 0;
      margin-left: 8px;
      padding: 2px 6px;
      font-size: 11px;
      color: #666;
      background: #f5f5f5;
      border-radius: 4px;
    }
  }

  .input-footer {
    text-align: center;
    font-size: 12px;
    color: $claude-text-muted;
    margin-top: 12px;
  }
}

// Workspace 面板过渡动画
.workspace-slide-enter-active,
.workspace-slide-leave-active {
  transition: all 0.3s ease;
}

.workspace-slide-enter,
.workspace-slide-leave-to {
  transform: translateX(100%);
  opacity: 0;
}

// Workspace 面板容器（需要添加）
.workspace-panel {
  width: 320px;
  flex-shrink: 0;
}
</style>
