/**
 * 预处理工具结果块
 * 将 <<< ... >>> 包裹的内容转换为自定义 HTML，
 * 自动区分 JSON 数据与纯文本描述，应用不同样式。
 */

function escapeHtml(str) {
  return str
    .replace(/&/g, '&amp;')
    .replace(/</g, '&lt;')
    .replace(/>/g, '&gt;')
    .replace(/"/g, '&quot;');
}

/**
 * @param {string} text 原始 markdown 文本
 * @returns {string} 替换后的文本（包含 HTML）
 */
export function processToolResultBlocks(text) {
  // 匹配 <<< 和 >>> 各自独占一行的块
  return text.replace(/^<<<\n([\s\S]*?)\n>>>$/gm, (match, content) => {
    const trimmed = content.trim();
    let isJson = false;
    let displayContent = trimmed;

    try {
      const parsed = JSON.parse(trimmed);
      isJson = true;
      displayContent = JSON.stringify(parsed, null, 2);
    } catch (e) {
      isJson = false;
    }

    if (isJson) {
      return `<div class="tool-result-box tool-result-json"><pre>${escapeHtml(displayContent)}</pre></div>`;
    } else {
      return `<div class="tool-result-box tool-result-text">${escapeHtml(trimmed)}</div>`;
    }
  });
}
