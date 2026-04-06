<template>
  <el-drawer
    :visible.sync="visible"
    direction="rtl"
    size="400px"
    :with-header="false"
    custom-class="config-drawer"
    @close="handleClose"
  >
    <div class="drawer-content">
      <div class="drawer-header">
        <h3>配置</h3>
        <i class="el-icon-close" @click="handleClose"></i>
      </div>

      <div class="drawer-body">
        <div class="drawer-section">
          <div class="section-header">
            <i class="el-icon-setting"></i>
            <span>工具选择</span>
            <el-tag size="mini" type="info">
              {{ selectedTools.length }} 已选
            </el-tag>
          </div>

          <div class="tool-search">
            <el-input
              v-model="searchKeyword"
              size="small"
              placeholder="搜索工具..."
              prefix-icon="el-icon-search"
              clearable
            />
          </div>

          <div v-if="loading" class="config-loading">
            <i class="el-icon-loading"></i>
            加载中...
          </div>
          <div
            v-else-if="filteredToolList.length === 0"
            class="config-empty"
          >
            <i class="el-icon-search"></i>
            <span>未找到匹配的工具</span>
          </div>
          <div v-else class="tool-categories">
            <div
              v-for="category in filteredToolList"
              :key="category.category"
              class="tool-category"
            >
              <div class="category-header">
                <span class="category-name">{{ category.category }}</span>
                <el-tag
                  size="mini"
                  :type="getConditionType(category.condition)"
                >
                  {{ getConditionLabel(category.condition) }}
                </el-tag>
              </div>
              <div class="tool-list">
                <el-tooltip
                  v-for="tool in category.toolList"
                  :key="tool.toolId"
                  placement="top"
                  :open-delay="500"
                  :disabled="!tool.description && !tool.desc"
                  effect="light"
                  popper-class="tool-tooltip-popper"
                >
                  <div slot="content" class="tool-detail-tooltip">
                    <div class="tooltip-title">{{ tool.toolName }}</div>
                    <div class="tooltip-desc">
                      {{ tool.description || tool.desc || '暂无详细描述' }}
                    </div>
                  </div>
                  <div
                    :class="[
                      'tool-item',
                      { selected: isToolSelected(tool.toolId) },
                    ]"
                    @click="handleToggleTool(tool)"
                  >
                    <div class="tool-avatar">
                      <img v-if="tool.avatar?.path" :src="tool.avatar.path" />
                      <i v-else class="el-icon-setting"></i>
                    </div>
                    <div class="tool-info">
                      <div class="tool-name">{{ tool.toolName }}</div>
                      <div class="tool-desc">{{ tool.desc }}</div>
                    </div>
                    <el-checkbox
                      :value="isToolSelected(tool.toolId)"
                      @click.native.stop
                      @change="handleToggleTool(tool)"
                    />
                  </div>
                </el-tooltip>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>
  </el-drawer>
</template>

<script>
export default {
  name: 'ConfigDrawer',
  props: {
    visible: {
      type: Boolean,
      default: false,
    },
    toolList: {
      type: Array,
      default: () => [],
    },
    selectedTools: {
      type: Array,
      default: () => [],
    },
    loading: {
      type: Boolean,
      default: false,
    },
  },
  data() {
    return {
      searchKeyword: '',
    };
  },
  computed: {
    filteredToolList() {
      if (!this.searchKeyword.trim()) {
        return this.toolList;
      }
      const keyword = this.searchKeyword.toLowerCase().trim();
      return this.toolList
        .map(category => {
          const filteredTools = category.toolList.filter(tool => {
            const name = (tool.toolName || '').toLowerCase();
            const desc = (tool.desc || '').toLowerCase();
            const description = (tool.description || '').toLowerCase();
            return (
              name.includes(keyword) ||
              desc.includes(keyword) ||
              description.includes(keyword)
            );
          });
          if (filteredTools.length === 0) return null;
          return {
            ...category,
            toolList: filteredTools,
          };
        })
        .filter(Boolean);
    },
  },
  methods: {
    handleClose() {
      this.$emit('update:visible', false);
      this.$emit('close');
    },

    isToolSelected(toolId) {
      return this.selectedTools.some(t => t.toolId === toolId);
    },

    handleToggleTool(tool) {
      this.$emit('toggle-tool', tool);
    },

    getConditionLabel(condition) {
      const labels = {
        none: '可选',
        optional: '推荐',
        required: '必选',
      };
      return labels[condition] || condition;
    },

    getConditionType(condition) {
      const types = {
        none: 'info',
        optional: 'warning',
        required: 'danger',
      };
      return types[condition] || 'info';
    },
  },
};
</script>

<style lang="scss">
.config-drawer {
  .drawer-content {
    height: 100%;
    display: flex;
    flex-direction: column;
  }

  .drawer-header {
    display: flex;
    align-items: center;
    justify-content: space-between;
    padding: 16px 20px;
    border-bottom: 1px solid #e5e5e5;

    h3 {
      margin: 0;
      font-size: 16px;
      font-weight: 500;
      color: #1a1a1a;
    }

    .el-icon-close {
      font-size: 18px;
      color: #999;
      cursor: pointer;
      transition: color 0.2s;

      &:hover {
        color: #10a37f;
      }
    }
  }

  .drawer-body {
    flex: 1;
    overflow-y: auto;
    padding: 16px 20px;
  }

  .drawer-section {
    margin-bottom: 24px;

    .section-header {
      display: flex;
      align-items: center;
      gap: 8px;
      margin-bottom: 16px;
      font-size: 14px;
      font-weight: 500;
      color: #1a1a1a;

      i {
        font-size: 16px;
        color: #10a37f;
      }
    }
  }

  .config-loading {
    text-align: center;
    color: #999;
    padding: 24px;
  }

  .config-empty {
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    padding: 32px;
    color: #999;

    i {
      font-size: 32px;
      margin-bottom: 8px;
    }

    span {
      font-size: 14px;
    }
  }

  .tool-categories {
    .tool-category {
      margin-bottom: 16px;

      .category-header {
        display: flex;
        align-items: center;
        justify-content: space-between;
        margin-bottom: 10px;
        padding-bottom: 8px;
        border-bottom: 1px solid #f0f0f0;

        .category-name {
          font-size: 13px;
          font-weight: 500;
          color: #1a1a1a;
        }
      }
    }
  }

  .tool-list {
    display: flex;
    flex-direction: column;
    gap: 6px;
  }

  .tool-item {
    display: flex;
    align-items: center;
    padding: 10px 12px;
    border-radius: 10px;
    cursor: pointer;
    transition: all 0.2s;
    border: 1px solid transparent;

    &:hover {
      background: #f5f7fa;
      border-color: #e4e7ed;
    }

    &.selected {
      background: rgba(16, 163, 127, 0.08);
      border-color: rgba(16, 163, 127, 0.2);
    }

    .tool-avatar {
      width: 36px;
      height: 36px;
      border-radius: 8px;
      margin-right: 12px;
      display: flex;
      align-items: center;
      justify-content: center;
      background: #f0f0f0;
      overflow: hidden;
      flex-shrink: 0;

      img {
        width: 100%;
        height: 100%;
        object-fit: cover;
      }

      i {
        font-size: 18px;
        color: #999;
      }
    }

    .tool-info {
      flex: 1;
      min-width: 0;

      .tool-name {
        font-size: 14px;
        font-weight: 500;
        color: #1a1a1a;
        margin-bottom: 2px;
      }

      .tool-desc {
        font-size: 12px;
        color: #666;
        overflow: hidden;
        text-overflow: ellipsis;
        white-space: nowrap;
      }
    }

    .el-checkbox {
      margin-left: 8px;
    }
  }

  .tool-search {
    margin-bottom: 16px;

    .el-input {
      .el-input__inner {
        border-radius: 8px;
      }
    }
  }
}

.tool-tooltip-popper {
  max-width: 360px !important;
  padding: 0 !important;
  border: 1px solid #e4e7ed !important;
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.12) !important;
  border-radius: 8px !important;
}

.tool-detail-tooltip {
  padding: 12px 14px;

  .tooltip-title {
    font-size: 14px;
    font-weight: 600;
    color: #1a1a1a;
    margin-bottom: 8px;
    padding-bottom: 8px;
    border-bottom: 1px solid #f0f0f0;
  }

  .tooltip-desc {
    font-size: 13px;
    color: #666;
    line-height: 1.6;
    white-space: pre-wrap;
    max-height: 200px;
    overflow-y: auto;
  }
}
</style>
