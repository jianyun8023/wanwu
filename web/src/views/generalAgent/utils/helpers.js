/**
 * 通用工具函数
 */

/**
 * 防抖函数
 * @param {Function} fn - 要防抖的函数
 * @param {number} delay - 延迟时间（毫秒）
 * @returns {Function} 防抖后的函数
 */
export function debounce(fn, delay = 100) {
  let timer = null;
  return function (...args) {
    if (timer) {
      clearTimeout(timer);
    }
    timer = setTimeout(() => {
      fn.apply(this, args);
      timer = null;
    }, delay);
  };
}

/**
 * 节流函数
 * @param {Function} fn - 要节流的函数
 * @param {number} interval - 执行间隔（毫秒）
 * @returns {Function} 节流后的函数
 */
export function throttle(fn, interval = 100) {
  let lastTime = 0;
  return function (...args) {
    const now = Date.now();
    if (now - lastTime >= interval) {
      fn.apply(this, args);
      lastTime = now;
    }
  };
}

/**
 * 格式化文件大小
 * @param {number} bytes - 字节数
 * @returns {string} 格式化后的字符串
 */
export function formatFileSize(bytes) {
  if (!bytes || bytes === 0) return '0 B';
  const k = 1024;
  const sizes = ['B', 'KB', 'MB', 'GB'];
  const i = Math.floor(Math.log(bytes) / Math.log(k));
  return Math.round((bytes / Math.pow(k, i)) * 100) / 100 + ' ' + sizes[i];
}

/**
 * 检测文件类型
 * @param {string} fileName - 文件名
 * @returns {string} 文件类型
 */
export function getFileType(fileName) {
  if (!fileName) return 'unsupported';
  const ext = fileName.split('.').pop().toLowerCase();

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
}

/**
 * 检测是否为图片文件
 * @param {Object} file - 文件对象
 * @returns {boolean} 是否为图片
 */
export function isImageFile(file) {
  const imageTypes = [
    'image/jpeg',
    'image/png',
    'image/gif',
    'image/webp',
    'image/bmp',
  ];
  const imageExts = ['jpg', 'jpeg', 'png', 'gif', 'webp', 'bmp'];

  if (file.type && imageTypes.includes(file.type)) {
    return true;
  }

  if (file.name) {
    const ext = file.name.split('.').pop().toLowerCase();
    return imageExts.includes(ext);
  }

  return false;
}

/**
 * 格式化持续时间
 * @param {number} ms - 毫秒数
 * @returns {string} 格式化后的时间字符串（如 "2m 30s" 或 "500ms"）
 */
export function formatDuration(ms) {
  if (ms === 0) {
    return '<1s';
  }
  if (ms < 1000) {
    return `${ms}ms`;
  }
  const seconds = Math.floor(ms / 1000);
  const minutes = Math.floor(seconds / 60);
  const secs = seconds % 60;
  if (minutes > 0) {
    return `${minutes}m ${secs}s`;
  }
  return `${secs}s`;
}

/**
 * 获取文件图标类名
 * @param {Object} file - 文件对象
 * @returns {string} Element UI 图标类名
 */
export function getFileIcon(file) {
  if (file.type === 'directory' || file.type === 'dir' || file.isDir) {
    return 'el-icon-folder';
  }

  const ext = file.name ? file.name.split('.').pop().toLowerCase() : '';
  const iconMap = {
    // 图片
    png: 'el-icon-picture',
    jpg: 'el-icon-picture',
    jpeg: 'el-icon-picture',
    gif: 'el-icon-picture',
    svg: 'el-icon-picture',
    webp: 'el-icon-picture',
    bmp: 'el-icon-picture',
    ico: 'el-icon-picture',
    // 视频
    mp4: 'el-icon-video-camera',
    webm: 'el-icon-video-camera',
    ogg: 'el-icon-video-camera',
    mov: 'el-icon-video-camera',
    m4v: 'el-icon-video-camera',
    avi: 'el-icon-video-camera',
    mkv: 'el-icon-video-camera',
    // 音频
    mp3: 'el-icon-headset',
    wav: 'el-icon-headset',
    m4a: 'el-icon-headset',
    flac: 'el-icon-headset',
    aac: 'el-icon-headset',
    // 文档
    pdf: 'el-icon-document',
    doc: 'el-icon-document',
    docx: 'el-icon-document',
    xls: 'el-icon-document',
    xlsx: 'el-icon-document',
    ppt: 'el-icon-document',
    pptx: 'el-icon-document',
    txt: 'el-icon-document',
    md: 'el-icon-document',
    html: 'el-icon-document',
    htm: 'el-icon-document',
    json: 'el-icon-document',
    js: 'el-icon-document',
    ts: 'el-icon-document',
    vue: 'el-icon-document',
    py: 'el-icon-document',
    java: 'el-icon-document',
    go: 'el-icon-document',
    css: 'el-icon-document',
    scss: 'el-icon-document',
    xml: 'el-icon-document',
    yaml: 'el-icon-document',
    yml: 'el-icon-document',
    sql: 'el-icon-document',
    sh: 'el-icon-document',
    // 压缩包
    zip: 'el-icon-files',
    rar: 'el-icon-files',
    tar: 'el-icon-files',
    gz: 'el-icon-files',
    '7z': 'el-icon-files',
  };

  return iconMap[ext] || 'el-icon-document';
}
