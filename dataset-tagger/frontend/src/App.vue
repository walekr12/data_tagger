<template>
  <div class="app-container h-screen flex flex-col bg-cyber-dark">
    <!-- 顶部工具栏 -->
    <header class="header glass-card m-2 mb-0 p-4 flex items-center justify-between">
      <div class="flex items-center gap-4">
        <div class="logo flex items-center gap-2">
          <div class="w-8 h-8 rounded-lg bg-gradient-to-br from-cyber-blue to-cyber-purple flex items-center justify-center">
            <span class="text-white font-bold text-lg">D</span>
          </div>
          <h1 class="text-xl font-bold bg-gradient-to-r from-cyber-blue to-cyber-purple bg-clip-text text-transparent">
            AI数据集打标器
          </h1>
        </div>
        
        <button @click="selectFolder" class="cyber-btn flex items-center gap-2">
          <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M3 7v10a2 2 0 002 2h14a2 2 0 002-2V9a2 2 0 00-2-2h-6l-2-2H5a2 2 0 00-2 2z" />
          </svg>
          导入文件夹
        </button>
        
        <span v-if="folderPath" class="text-gray-400 text-sm truncate max-w-xs">
          {{ folderPath }}
        </span>
      </div>
      
      <div class="flex items-center gap-3">
        <button v-if="items.length > 0" @click="showBatchPanel = !showBatchPanel" 
                class="cyber-btn" :class="{ 'neon-glow': showBatchPanel }">
          批量操作
        </button>
        
        <!-- 刷新统计按钮 -->
        <button v-if="items.length > 0" @click="refreshTagStats" class="cyber-btn flex items-center gap-1">
          <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15" />
          </svg>
          刷新统计
        </button>
        
        <!-- 保存全部按钮（始终显示，有修改时高亮） -->
        <button v-if="items.length > 0" @click="saveAllChanges" 
                class="cyber-btn flex items-center gap-1"
                :class="{ 'cyber-btn-primary neon-glow': hasModifiedItems }">
          <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 13l4 4L19 7" />
          </svg>
          保存全部 {{ hasModifiedItems ? `(${modifiedCount})` : '' }}
        </button>
      </div>
    </header>

    <!-- 主内容区 -->
    <div class="main-content flex-1 flex overflow-hidden m-2 gap-2">
      <!-- 左侧标签面板 -->
      <aside v-if="tags.length > 0" class="tags-panel glass-card w-72 flex flex-col overflow-hidden">
        <div class="p-4 border-b border-cyber-blue/20">
          <h2 class="text-lg font-semibold text-cyber-blue mb-2">标签词频</h2>
          <input v-model="tagSearch" type="text" placeholder="搜索标签..." class="cyber-input text-sm">
        </div>
        
        <div class="flex-1 overflow-y-auto p-4">
          <div class="flex flex-wrap gap-2">
            <button v-for="tag in filteredTags" :key="tag.tag"
                    @click="filterByTag(tag.tag)"
                    class="tag-pill"
                    :class="[
                      selectedTag === tag.tag ? 'tag-pill-purple neon-glow-purple' : 'tag-pill-blue',
                      getTagSizeClass(tag.count)
                    ]">
              <span>{{ tag.tag }}</span>
              <span class="ml-1 opacity-60">({{ tag.count }})</span>
            </button>
          </div>
        </div>
        
        <div v-if="selectedTag" class="p-3 border-t border-cyber-blue/20">
          <button @click="clearTagFilter" class="cyber-btn w-full text-sm">
            清除筛选
          </button>
        </div>
      </aside>

      <!-- 中间网格预览 -->
      <main class="grid-panel glass-card flex-1 flex flex-col overflow-hidden">
        <!-- 批量操作面板 -->
        <div v-if="showBatchPanel" class="batch-panel p-4 border-b border-cyber-blue/20 fade-in">
          <div class="flex items-center gap-4 mb-4">
            <label class="flex items-center gap-2 cursor-pointer">
              <input type="checkbox" v-model="selectAll" @change="toggleSelectAll" 
                     class="w-4 h-4 rounded border-cyber-blue/50 bg-transparent">
              <span class="text-sm">全选 ({{ selectedItems.length }}/{{ displayItems.length }})</span>
            </label>
            
            <label class="flex items-center gap-2 cursor-pointer">
              <input type="checkbox" v-model="useRegex" class="w-4 h-4 rounded">
              <span class="text-sm">使用正则</span>
            </label>
          </div>
          
          <div class="grid grid-cols-3 gap-4">
            <!-- 添加标签 -->
            <div class="space-y-2">
              <label class="text-sm text-gray-400">添加标签</label>
              <input v-model="batchAddTagValue" type="text" placeholder="输入标签..." class="cyber-input text-sm">
              <div class="flex gap-2">
                <button @click="batchAddTag('prepend')" class="cyber-btn text-xs flex-1">添加到开头</button>
                <button @click="batchAddTag('append')" class="cyber-btn text-xs flex-1">添加到末尾</button>
              </div>
            </div>
            
            <!-- 删除标签 -->
            <div class="space-y-2">
              <label class="text-sm text-gray-400">删除标签</label>
              <input v-model="batchRemoveTagValue" type="text" placeholder="输入要删除的标签..." class="cyber-input text-sm">
              <button @click="batchRemoveTag" class="cyber-btn cyber-btn-danger text-xs w-full">删除匹配标签</button>
            </div>
            
            <!-- 替换标签 -->
            <div class="space-y-2">
              <label class="text-sm text-gray-400">替换标签</label>
              <input v-model="batchReplaceOld" type="text" placeholder="原标签..." class="cyber-input text-sm">
              <input v-model="batchReplaceNew" type="text" placeholder="新标签..." class="cyber-input text-sm">
              <button @click="batchReplaceTag" class="cyber-btn text-xs w-full">替换</button>
            </div>
          </div>
        </div>

        <!-- 网格视图 -->
        <div v-if="loading" class="flex-1 flex items-center justify-center">
          <div class="text-center">
            <div class="w-16 h-16 border-4 border-cyber-blue/30 border-t-cyber-blue rounded-full animate-spin mx-auto mb-4"></div>
            <p class="text-cyber-blue">{{ loadingMessage }}</p>
          </div>
        </div>
        
        <div v-else-if="displayItems.length === 0" class="flex-1 flex items-center justify-center">
          <div class="text-center text-gray-500">
            <svg class="w-24 h-24 mx-auto mb-4 opacity-30" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="1" d="M4 16l4.586-4.586a2 2 0 012.828 0L16 16m-2-2l1.586-1.586a2 2 0 012.828 0L20 14m-6-6h.01M6 20h12a2 2 0 002-2V6a2 2 0 00-2-2H6a2 2 0 00-2 2v12a2 2 0 002 2z" />
            </svg>
            <p class="text-lg">请导入数据集文件夹</p>
            <p class="text-sm mt-2">支持图片 (jpg, png, gif, webp) 和视频 (mp4, avi, mov)</p>
          </div>
        </div>
        
        <div v-else class="flex-1 overflow-y-auto p-4">
          <div class="grid grid-cols-2 md:grid-cols-3 lg:grid-cols-4 xl:grid-cols-5 gap-4">
            <div v-for="item in pagedItems" :key="item.id"
                 class="media-card-vertical glass-card flex flex-col"
                 :class="{ 
                   'selected-card': item.selected,
                   'highlight-match': selectedTag && item.tags.includes(selectedTag)
                 }">
              
              <!-- 顶部图片区域 -->
              <div class="relative aspect-square cursor-pointer group" @click="openEditor(item)">
                <!-- 选择框 -->
                <div v-if="showBatchPanel" class="absolute top-2 left-2 z-20" @click.stop>
                  <input type="checkbox" v-model="item.selected" 
                         class="w-5 h-5 rounded border-cyber-blue bg-cyber-dark/80 cursor-pointer">
                </div>
                
                <!-- 视频��识 -->
                <span v-if="item.isVideo" class="absolute top-2 right-2 z-10 text-xs px-2 py-1 rounded bg-cyber-purple/80 text-white font-bold">
                  VIDEO
                </span>
                
                <!-- 修改标识 -->
                <span v-if="item.modified" class="absolute top-2 right-2 z-10 w-3 h-3 rounded-full bg-cyber-yellow animate-pulse"></span>
                
                <!-- 缩略图 -->
                <img v-if="item.thumbnailData" 
                     :src="item.thumbnailData" 
                     :alt="item.id" 
                     loading="lazy"
                     class="w-full h-full object-cover rounded-t-lg">
                <div v-else class="w-full h-full bg-cyber-darker flex items-center justify-center rounded-t-lg">
                  <div class="text-center">
                    <svg class="w-10 h-10 text-gray-600 mx-auto animate-pulse" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                      <path stroke-linecap="round" stroke-linejoin="round" stroke-width="1" d="M4 16l4.586-4.586a2 2 0 012.828 0L16 16m-2-2l1.586-1.586a2 2 0 012.828 0L20 14m-6-6h.01M6 20h12a2 2 0 002-2V6a2 2 0 00-2-2H6a2 2 0 00-2 2v12a2 2 0 002 2z" />
                    </svg>
                    <span class="text-xs text-gray-500 mt-1">加载中...</span>
                  </div>
                </div>
                
                <!-- 悬停遮罩 -->
                <div class="absolute inset-0 bg-cyber-blue/10 opacity-0 group-hover:opacity-100 transition-opacity flex items-center justify-center">
                  <span class="px-3 py-1 bg-cyber-blue/80 rounded text-white text-sm">点击编辑</span>
                </div>
              </div>
              
              <!-- 底部标签区域 - 可编辑框 -->
              <div class="tags-box p-2 border-t border-cyber-blue/20 min-h-[80px] max-h-[120px] overflow-y-auto">
                <div v-if="editingCardId === item.id" class="h-full">
                  <textarea 
                    v-model="editingCardTags"
                    @blur="saveCardTags(item)"
                    @keydown.enter.ctrl="saveCardTags(item)"
                    class="w-full h-full bg-transparent text-xs text-gray-300 resize-none border-none outline-none p-0"
                    placeholder="输入标签，逗号分隔..."
                    ref="cardTagInput"
                  ></textarea>
                </div>
                <div v-else @click.stop="startEditCard(item)" class="cursor-text h-full">
                  <!-- 高亮显示匹配的词 -->
                  <div v-if="item.rawTags" class="text-xs text-gray-300 leading-relaxed">
                    <span v-html="highlightText(item.rawTags)"></span>
                  </div>
                  <div v-else class="text-xs text-gray-500 italic">点击添加标签...</div>
                </div>
              </div>
            </div>
          </div>
        </div>

        <!-- 分页 -->
        <div v-if="totalPages > 1" class="pagination p-4 border-t border-cyber-blue/20 flex items-center justify-center gap-2">
          <button @click="currentPage = 1" :disabled="currentPage === 1" class="cyber-btn text-sm px-3">
            «
          </button>
          <button @click="currentPage--" :disabled="currentPage === 1" class="cyber-btn text-sm px-3">
            ‹
          </button>
          
          <span class="px-4 text-gray-400">
            {{ currentPage }} / {{ totalPages }}
          </span>
          
          <button @click="currentPage++" :disabled="currentPage === totalPages" class="cyber-btn text-sm px-3">
            ›
          </button>
          <button @click="currentPage = totalPages" :disabled="currentPage === totalPages" class="cyber-btn text-sm px-3">
            »
          </button>
          
          <select v-model="pageSize" class="cyber-input w-20 text-sm ml-4">
            <option :value="12">12</option>
            <option :value="24">24</option>
            <option :value="48">48</option>
            <option :value="96">96</option>
          </select>
        </div>
      </main>
    </div>

    <!-- 状态栏 -->
    <footer class="status-bar">
      <div class="flex items-center gap-2">
        <span :class="statusClass">●</span>
        <span>{{ statusMessage }}</span>
      </div>
      
      <div class="flex-1"></div>
      
      <div v-if="items.length > 0" class="flex items-center gap-4 text-gray-400">
        <span>图片: {{ totalImages }}</span>
        <span>视频: {{ totalVideos }}</span>
        <span>总计: {{ items.length }}</span>
      </div>
    </footer>

    <!-- 编辑器模态框 -->
    <div v-if="editingItem" class="modal-overlay" @click.self="closeEditor">
      <div class="modal-content w-[90vw] h-[85vh] flex">
        <!-- 左侧媒体预览 -->
        <div class="media-preview flex-1 bg-black flex items-center justify-center p-4">
          <img v-if="!editingItem.isVideo && previewData" 
               :src="previewData" 
               class="max-w-full max-h-full object-contain rounded-lg">
          <video v-else-if="editingItem.isVideo && previewData" 
                 :src="previewData" 
                 controls 
                 class="max-w-full max-h-full rounded-lg">
          </video>
          <div v-else class="text-gray-500">加载中...</div>
        </div>
        
        <!-- 右侧编辑面板 -->
        <div class="edit-panel w-96 flex flex-col border-l border-cyber-blue/20 bg-cyber-dark">
          <div class="p-4 border-b border-cyber-blue/20 flex items-center justify-between">
            <h3 class="text-lg font-semibold text-cyber-blue">编辑标签</h3>
            <button @click="closeEditor" class="text-gray-400 hover:text-white">
              <svg class="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
              </svg>
            </button>
          </div>
          
          <div class="p-4 text-sm text-gray-400 border-b border-cyber-blue/20">
            <p class="truncate">{{ getFileName(editingItem.mediaPath) }}</p>
          </div>
          
          <div class="flex-1 p-4 overflow-y-auto">
            <label class="text-sm text-gray-400 mb-2 block">标签内容 (逗号分隔)</label>
            <textarea v-model="editingTags" 
                      class="cyber-input h-48 resize-none font-mono text-sm"
                      placeholder="输入标签，用逗号分隔..."></textarea>
            
            <div class="mt-4">
              <label class="text-sm text-gray-400 mb-2 block">当前标签</label>
              <div class="flex flex-wrap gap-2">
                <span v-for="(tag, idx) in parsedEditingTags" :key="idx"
                      class="tag-pill tag-pill-blue">
                  {{ tag }}
                  <button @click="removeEditingTag(idx)" class="ml-1 hover:text-red-400">×</button>
                </span>
              </div>
            </div>
          </div>
          
          <div class="p-4 border-t border-cyber-blue/20 flex gap-2">
            <button @click="closeEditor" class="cyber-btn flex-1">取消</button>
            <button @click="saveCurrentItem" class="cyber-btn cyber-btn-primary flex-1">保存</button>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script>
export default {
  name: 'App',
  data() {
    return {
      // 数据
      folderPath: '',
      items: [],
      tags: [],
      totalImages: 0,
      totalVideos: 0,
      
      // UI状态
      loading: false,
      loadingMessage: '',
      statusMessage: '就绪',
      statusType: 'success',
      
      // 筛选
      selectedTag: null,
      tagSearch: '',
      
      // 分页
      currentPage: 1,
      pageSize: 24,
      
      // 批量操作
      showBatchPanel: false,
      selectAll: false,
      useRegex: false,
      batchAddTagValue: '',
      batchRemoveTagValue: '',
      batchReplaceOld: '',
      batchReplaceNew: '',
      
      // 编辑器
      editingItem: null,
      editingTags: '',
      previewData: null,
      
      // 卡片内编辑
      editingCardId: null,
      editingCardTags: ''
    }
  },
  
  computed: {
    filteredTags() {
      if (!this.tagSearch) return this.tags
      const search = this.tagSearch.toLowerCase()
      return this.tags.filter(t => t.tag.toLowerCase().includes(search))
    },
    
    displayItems() {
      if (this.selectedTag) {
        // 使用子串匹配，因为标签是共同短语
        return this.items.filter(item => item.rawTags && item.rawTags.includes(this.selectedTag))
      }
      return this.items
    },
    
    pagedItems() {
      const start = (this.currentPage - 1) * this.pageSize
      return this.displayItems.slice(start, start + this.pageSize)
    },
    
    totalPages() {
      return Math.ceil(this.displayItems.length / this.pageSize)
    },
    
    selectedItems() {
      return this.displayItems.filter(item => item.selected)
    },
    
    hasModifiedItems() {
      return this.items.some(item => item.modified)
    },
    
    modifiedCount() {
      return this.items.filter(item => item.modified).length
    },
    
    parsedEditingTags() {
      if (!this.editingTags) return []
      return this.editingTags.split(',').map(t => t.trim()).filter(t => t)
    },
    
    statusClass() {
      switch(this.statusType) {
        case 'success': return 'status-success'
        case 'error': return 'status-error'
        case 'loading': return 'status-loading animate-pulse'
        default: return 'text-gray-400'
      }
    }
  },
  
  watch: {
    currentPage() {
      // 翻页时加载新页面的缩略图
      this.$nextTick(() => this.loadVisibleThumbnails())
    },
    pageSize() {
      this.currentPage = 1
      this.$nextTick(() => this.loadVisibleThumbnails())
    }
  },
  
  methods: {
    async selectFolder() {
      try {
        const path = await window.go.main.App.SelectFolder()
        if (path) {
          this.folderPath = path
          await this.scanFolder(path)
        }
      } catch (err) {
        this.setStatus('选择文件夹失败: ' + err, 'error')
      }
    },
    
    async scanFolder(path) {
      this.loading = true
      this.loadingMessage = '正在扫描文件夹...'
      this.setStatus('扫描中...', 'loading')
      
      try {
        const result = await window.go.main.App.ScanFolder(path)
        
        if (result.success) {
          this.items = result.items
          this.tags = result.tags
          this.totalImages = result.totalImages
          this.totalVideos = result.totalVideos
          this.currentPage = 1
          
          this.setStatus(result.message, 'success')
          
          // 开始加载缩略图
          await this.loadVisibleThumbnails()
        } else {
          this.setStatus(result.message, 'error')
        }
      } catch (err) {
        this.setStatus('扫描失败: ' + err, 'error')
      } finally {
        this.loading = false
      }
    },
    
    async loadVisibleThumbnails() {
      this.loadingMessage = '正在生成缩略图...'
      
      // 并行加载缩略图，每批4个
      const itemsToLoad = this.pagedItems.filter(item => !item.thumbnailData)
      const batchSize = 4
      
      for (let i = 0; i < itemsToLoad.length; i += batchSize) {
        const batch = itemsToLoad.slice(i, i + batchSize)
        
        await Promise.all(batch.map(async (item) => {
          await this.loadThumbnailWithRetry(item, 5) // 最多重试5次
        }))
        
        // 更新进度
        this.loadingMessage = `正在生成缩略图... (${Math.min(i + batchSize, itemsToLoad.length)}/${itemsToLoad.length})`
      }
    },
    
    // 带重试的缩略图加载
    async loadThumbnailWithRetry(item, maxRetries = 5, delay = 1000) {
      for (let attempt = 1; attempt <= maxRetries; attempt++) {
        try {
          const thumb = await window.go.main.App.GetThumbnail(item.mediaPath, item.isVideo)
          if (thumb) {
            const idx = this.items.findIndex(i => i.id === item.id)
            if (idx !== -1) {
              this.items[idx].thumbnailData = thumb
            }
            return true // 加载成功
          }
        } catch (err) {
          console.warn(`缩略图加载失败 (尝试 ${attempt}/${maxRetries}):`, item.mediaPath, err)
        }
        
        // 如果不是最后一次尝试，等待后重试
        if (attempt < maxRetries) {
          await new Promise(resolve => setTimeout(resolve, delay))
        }
      }
      
      console.error('缩略图加载最终失败:', item.mediaPath)
      // 设置一个失败标记，显示占位图但允许后续重试
      const idx = this.items.findIndex(i => i.id === item.id)
      if (idx !== -1) {
        this.items[idx].thumbnailFailed = true
      }
      return false
    },
    
    // 手动重试加载失败的缩略图
    async retryFailedThumbnails() {
      const failedItems = this.pagedItems.filter(item => item.thumbnailFailed && !item.thumbnailData)
      if (failedItems.length === 0) return
      
      this.setStatus(`正在重新加载 ${failedItems.length} 个失败的缩略图...`, 'loading')
      
      for (const item of failedItems) {
        item.thumbnailFailed = false
        await this.loadThumbnailWithRetry(item, 3)
      }
      
      const stillFailed = failedItems.filter(item => item.thumbnailFailed).length
      if (stillFailed > 0) {
        this.setStatus(`${failedItems.length - stillFailed} 个加载成功，${stillFailed} 个仍然失败`, 'error')
      } else {
        this.setStatus('所有缩略图加载成功', 'success')
      }
    },
    
    filterByTag(tag) {
      if (this.selectedTag === tag) {
        this.selectedTag = null
      } else {
        this.selectedTag = tag
        this.currentPage = 1
      }
    },
    
    clearTagFilter() {
      this.selectedTag = null
      this.currentPage = 1
    },
    
    getTagSizeClass(count) {
      const maxCount = this.tags.length > 0 ? this.tags[0].count : 1
      const ratio = count / maxCount
      if (ratio > 0.7) return 'text-base font-bold'
      if (ratio > 0.4) return 'text-sm font-medium'
      return 'text-xs'
    },
    
    toggleSelectAll() {
      this.displayItems.forEach(item => {
        item.selected = this.selectAll
      })
    },
    
    async batchAddTag(position) {
      if (!this.batchAddTagValue.trim()) return
      if (this.selectedItems.length === 0) {
        this.setStatus('请先选择项目', 'error')
        return
      }
      
      try {
        const ids = this.selectedItems.map(i => i.id)
        await window.go.main.App.BatchAddTag(ids, this.batchAddTagValue.trim(), position)
        
        // 更新本地数据
        this.selectedItems.forEach(item => {
          if (position === 'prepend') {
            item.rawTags = this.batchAddTagValue.trim() + ', ' + item.rawTags
          } else {
            item.rawTags = item.rawTags + ', ' + this.batchAddTagValue.trim()
          }
          item.tags = this.parseTags(item.rawTags)
          item.modified = true
        })
        
        this.setStatus(`已为 ${ids.length} 个项目添加标签`, 'success')
        this.batchAddTagValue = ''
      } catch (err) {
        this.setStatus('批量添加失败: ' + err, 'error')
      }
    },
    
    async batchRemoveTag() {
      if (!this.batchRemoveTagValue.trim()) return
      if (this.selectedItems.length === 0) {
        this.setStatus('请先选择项目', 'error')
        return
      }
      
      try {
        const ids = this.selectedItems.map(i => i.id)
        await window.go.main.App.BatchRemoveTag(ids, this.batchRemoveTagValue.trim(), this.useRegex)
        
        // 刷新数据
        await this.refreshItems()
        this.setStatus(`已从 ${ids.length} 个项目删除匹配标签`, 'success')
        this.batchRemoveTagValue = ''
      } catch (err) {
        this.setStatus('批量删除失败: ' + err, 'error')
      }
    },
    
    async batchReplaceTag() {
      if (!this.batchReplaceOld.trim() || !this.batchReplaceNew.trim()) return
      if (this.selectedItems.length === 0) {
        this.setStatus('请先选择项目', 'error')
        return
      }
      
      try {
        const ids = this.selectedItems.map(i => i.id)
        await window.go.main.App.BatchReplaceTag(ids, this.batchReplaceOld.trim(), this.batchReplaceNew.trim(), this.useRegex)
        
        await this.refreshItems()
        this.setStatus(`已替换 ${ids.length} 个项目的标签`, 'success')
        this.batchReplaceOld = ''
        this.batchReplaceNew = ''
      } catch (err) {
        this.setStatus('批量替换失败: ' + err, 'error')
      }
    },
    
    async refreshItems() {
      const result = await window.go.main.App.GetItems()
      // 保留缩略图数据
      result.forEach(item => {
        const existing = this.items.find(i => i.id === item.id)
        if (existing) {
          item.thumbnailData = existing.thumbnailData
        }
      })
      this.items = result
    },
    
    async openEditor(item) {
      this.editingItem = item
      this.editingTags = item.rawTags
      this.previewData = null
      
      // 加载完整预览
      try {
        this.previewData = await window.go.main.App.ReadMediaFile(item.mediaPath)
      } catch (err) {
        console.error('加载预览失败:', err)
      }
    },
    
    closeEditor() {
      this.editingItem = null
      this.editingTags = ''
      this.previewData = null
    },
    
    removeEditingTag(idx) {
      const tags = this.parsedEditingTags
      tags.splice(idx, 1)
      this.editingTags = tags.join(', ')
    },
    
    async saveCurrentItem() {
      try {
        await window.go.main.App.SaveTags(this.editingItem.id, this.editingTags)
        
        // 更新本地数据
        const idx = this.items.findIndex(i => i.id === this.editingItem.id)
        if (idx !== -1) {
          this.items[idx].rawTags = this.editingTags
          this.items[idx].tags = this.parseTags(this.editingTags)
          this.items[idx].modified = false
        }
        
        this.setStatus('保存成功', 'success')
        this.closeEditor()
        
        // 刷新标签统计
        await this.updateTagStats()
      } catch (err) {
        this.setStatus('保存失败: ' + err, 'error')
      }
    },
    
    async saveAllChanges() {
      const modifiedItems = this.items.filter(i => i.modified)
      if (modifiedItems.length === 0) return
      
      this.setStatus('保存中...', 'loading')
      
      try {
        for (const item of modifiedItems) {
          await window.go.main.App.SaveTags(item.id, item.rawTags)
          item.modified = false
        }
        
        this.setStatus(`已保存 ${modifiedItems.length} 个文件`, 'success')
      } catch (err) {
        this.setStatus('保存失败: ' + err, 'error')
      }
    },
    
    async updateTagStats() {
      // 重新统计标签
      const tagFreq = {}
      this.items.forEach(item => {
        item.tags.forEach(tag => {
          tagFreq[tag] = (tagFreq[tag] || 0) + 1
        })
      })
      
      this.tags = Object.entries(tagFreq)
        .map(([tag, count]) => ({ tag, count }))
        .sort((a, b) => b.count - a.count)
    },
    
    // 刷新统计 - 调用后端重新分析共同短语
    async refreshTagStats() {
      this.setStatus('正在重新统计标签...', 'loading')
      
      try {
        // 调用后端重新分析
        const result = await window.go.main.App.RefreshTagStats()
        
        if (result && result.tags) {
          this.tags = result.tags
          this.setStatus(`标签统计已刷新，共 ${this.tags.length} 个共同短语`, 'success')
        } else {
          // 如果后端没有这个方法，使用前端统计
          await this.updateTagStats()
          this.setStatus(`标签统计已刷新，共 ${this.tags.length} 个标签`, 'success')
        }
      } catch (err) {
        // 如果后端方法不存在，使用前端统计
        console.warn('后端刷新失败，使用前端统计:', err)
        await this.updateTagStats()
        this.setStatus(`标签统计已刷新，共 ${this.tags.length} 个标签`, 'success')
      }
    },
    
    parseTags(content) {
      if (!content) return []
      return content.split(',').map(t => t.trim()).filter(t => t)
    },
    
    getFileName(path) {
      return path.split(/[/\\]/).pop()
    },
    
    setStatus(message, type = 'success') {
      this.statusMessage = message
      this.statusType = type
    },
    
    // 卡片内直接编辑
    startEditCard(item) {
      this.editingCardId = item.id
      this.editingCardTags = item.rawTags
      this.$nextTick(() => {
        const input = this.$refs.cardTagInput
        if (input && input[0]) {
          input[0].focus()
        }
      })
    },
    
    async saveCardTags(item) {
      if (this.editingCardId !== item.id) return
      
      const newTags = this.editingCardTags.trim()
      const idx = this.items.findIndex(i => i.id === item.id)
      
      if (idx !== -1 && newTags !== item.rawTags) {
        this.items[idx].rawTags = newTags
        this.items[idx].tags = this.parseTags(newTags)
        this.items[idx].modified = true
        this.setStatus('标签已修改，请点击保存全部', 'success')
      }
      
      this.editingCardId = null
      this.editingCardTags = ''
    },
    
    // 高亮显示选中的标签/短语
    highlightText(text) {
      if (!text) return ''
      if (!this.selectedTag) {
        // 没有选中标签时，直接返回转义后的文本
        return this.escapeHtml(text)
      }
      
      // 转义HTML特殊字符
      const escapedText = this.escapeHtml(text)
      const escapedTag = this.escapeHtml(this.selectedTag)
      
      // 使用正则进行全局替换，高亮所有匹配项
      const regex = new RegExp(this.escapeRegex(escapedTag), 'gi')
      return escapedText.replace(regex, '<mark class="highlight-tag">$&</mark>')
    },
    
    escapeHtml(text) {
      const div = document.createElement('div')
      div.textContent = text
      return div.innerHTML
    },
    
    escapeRegex(string) {
      return string.replace(/[.*+?^${}()|[\]\\]/g, '\\$&')
    }
  }
}
</script>

<style scoped>
.app-container {
  background: linear-gradient(135deg, #0a0a0f 0%, #0f0f1a 50%, #0a0a0f 100%);
}

.header::before {
  content: '';
  position: absolute;
  top: 0;
  left: 0;
  right: 0;
  height: 1px;
  background: linear-gradient(90deg, transparent, var(--cyber-blue), transparent);
}
</style>
