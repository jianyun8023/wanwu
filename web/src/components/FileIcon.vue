<template>
  <svg class="icon-file" aria-hidden="true" :style="{ fontSize: size }">
    <use :xlink:href="iconHref"></use>
  </svg>
</template>

<!--  Iconfont 完整图标清单
 * * --- 基础与文件夹 ---
 * icon-md        : Markdown 文件
 * icon-unknown   : 未知文件类型
 * icon-dir       : 文件夹 (关闭)
 * icon-diropen   : 文件夹 (打开)
 * * --- 办公文档 ---
 * icon-doc       : Word 文档 (.doc)
 * icon-docx      : Word 文档 (.docx)
 * icon-xls       : Excel 表格 (.xls)
 * icon-xlsx      : Excel 表格 (.xlsx)
 * icon-csv       : CSV 纯文本表格
 * icon-pdf       : PDF 文档
 * icon-ppt       : PPT 演示文稿 (.ppt)
 * icon-pptx      : PPT 演示文稿 (.pptx)
 * icon-txt       : 纯文本文件
 * * --- 多媒体 (音频/视频/图片) ---
 * icon-aac       : AAC 音频
 * icon-flac      : FLAC 无损音频
 * icon-m4a       : M4A 音频
 * icon-midi      : MIDI 音乐序列
 * icon-mp3       : MP3 音频
 * icon-ogg       : OGG 格式音频
 * icon-wav       : WAV 音频
 * icon-mp2       : MP2 音频
 * icon-mp4       : MP4 视频
 * icon-mkv       : MKV 视频
 * icon-ogv       : OGV 视频
 * icon-webm      : WEBM 视频
 * icon-gif       : GIF 动图
 * icon-jpg       : JPG 图片
 * icon-png       : PNG 图片
 * icon-webp      : WEBP 图片
 * icon-svg       : SVG 矢量图
 * * --- 开发与压缩包 ---
 * icon-js        : JavaScript 脚本
 * icon-html      : HTML 页面
 * icon-php       : PHP 脚本
 * icon-asp       : ASP 脚本
 * icon-zip       : ZIP 压缩包
 * icon-rar       : RAR 压缩包
 */
-->
<script>
export default {
  name: 'FileIcon',
  props: {
    type: {
      type: String,
      default: 'unknown',
    },
    size: {
      type: String,
      default: '1em',
    },
  },
  computed: {
    iconHref() {
      return `#icon-${this.type}`;
    },
  },
};
</script>

<style lang="scss" scoped>
.icon-file {
  width: 1em;
  height: 1em;
  vertical-align: -0.15em;
  fill: currentColor;
  overflow: hidden;
}
</style>
