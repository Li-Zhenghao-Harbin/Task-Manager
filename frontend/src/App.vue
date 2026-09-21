<script setup lang="ts">
import { computed, defineComponent, h, onMounted, reactive, ref, watch } from 'vue'
import { CreateAgentApp, CreateTask, CreateUnit, DeleteAgentApp, DeleteTask, DeleteUnit, ExportData, GetDefaultUnitID, GetSettings, ImportData, ListAgentApps, ListTasks, ListUnits, SetDefaultUnit, UpdateAgentApp, UpdateSettings, UpdateTask, UpdateTaskStatus, UpdateUnit } from '../wailsjs/go/main/App'
import { EventsOn } from '../wailsjs/runtime/runtime'
import appLogo from './assets/task-manager-logo.svg'

type Status = 'in_progress' | 'completed' | 'cancelled'
type Kind = 'main' | 'branch' | 'child'
type TaskType = 'code' | 'document' | 'support'
type OutputKey = 'newFeatures' | 'modifiedFeatures' | 'applications' | 'skills' | 'newDocuments' | 'modifiedDocuments' | 'supports'
type Output = { result: string } & Record<OutputKey, number>
type Task = { id:string; name:string; kind:Kind; parentId:string; issuer:string; unit:string; publishDate:string; dueDate:string; taskType:TaskType; status:Status; completedAt:string; agentId:string; createdAt:string; updatedAt:string; output:Output }
type Unit = { id:string; name:string }
type AgentApp = { id:string; name:string; model:string; projectIds:string[]; available:boolean; resetAt:string; createdAt:string; updatedAt:string }

const OutputStepper = defineComponent({
  props: { modelValue: { type: Number, required: true } },
  emits: ['update:modelValue'],
  setup(props, { emit }) {
    const set = (value: number) => emit('update:modelValue', Math.max(0, Number(value) || 0))
    return () => h('span', { class: 'stepper' }, [
      h('button', { type: 'button', onClick: () => set(props.modelValue - 1) }, '−'),
      h('input', { type: 'text', inputmode: 'numeric', value: props.modelValue, onInput: (e: Event) => set(Number((e.target as HTMLInputElement).value)) }),
      h('button', { type: 'button', onClick: () => set(props.modelValue + 1) }, '＋'),
    ])
  },
})

const tasks = ref<Task[]>([])
const units = ref<Unit[]>([])
const agents = ref<AgentApp[]>([])
const selected = ref<Task | null>(null)
const selectedAgent = ref<AgentApp | null>(null)
const defaultUnitID = ref('')
const section = ref<'tasks' | 'output' | 'agent' | 'units'>('tasks')
const outputFilter = ref<'all' | TaskType>('all')
const showCreate = ref(false)
const showAgentCreate = ref(false)
const showAbout = ref(false)
const showSettings = ref(false)
const message = ref('')
const newUnit = ref('')
const query = ref('')
const status = ref<'all' | Status>('all')
const kindFilter = ref<'all' | Kind>('all')
const typeFilter = ref<'all' | TaskType>('all')
const viewMode = ref<'list' | 'calendar'>('list')
const calendarMonth = ref(new Date(new Date().getFullYear(), new Date().getMonth(), 1))
const settings = reactive({ backgroundColor: '#f7faff', autoStart: false })
const contextTask = ref<Task | null>(null)
const contextMenu = reactive({ visible: false, x: 0, y: 0 })
const confirmation = reactive({ visible: false, title: '', content: '', confirmText: '确认', action: null as null | (() => Promise<void>) })

// ---- 右侧面板即时保存 ----
// 面板里的每次修改都会自动落盘，因此不需要保存按钮，也不显示保存状态提示。
// 改动先入队再防抖冲刷；savedSignatures 保存“已落盘内容”的签名，
// 用来识别服务端回写与切换选中项，避免保存回环。
const dirtyTasks = new Map<string, Task>()
const dirtyAgents = new Map<string, AgentApp>()
const savedSignatures = new Map<string, string>()
const SAVE_DEBOUNCE = 500
let saveTimer: ReturnType<typeof setTimeout> | undefined

const today = () => new Date().toISOString().slice(0, 10)
const emptyOutput = (): Output => ({ result: '', newFeatures: 0, modifiedFeatures: 0, applications: 0, skills: 0, newDocuments: 0, modifiedDocuments: 0, supports: 0 })
const draft = reactive({ name: '', kind: 'main' as Kind, parentId: '', issuer: '', unit: '', publishDate: today(), dueDate: today(), taskType: 'code' as TaskType, agentId: '' })
const agentDraft = reactive({ name: '', model: '', projectIds: [] as string[], available: true, resetAt: '' })
const statusText: Record<Status, string> = { in_progress: '进行中', completed: '已完成', cancelled: '已取消' }
const kindText: Record<Kind, string> = { main: '主线任务', branch: '支线任务', child: '子任务' }
const outputLabels: Record<OutputKey, string> = { newFeatures: '新增功能', modifiedFeatures: '修改功能', applications: '应用程序', newDocuments: '新增文档', modifiedDocuments: '修改文档', skills: 'Skills', supports: '实施' }
const typeKeys: Record<TaskType, OutputKey[]> = { code: ['newFeatures', 'modifiedFeatures', 'applications'], document: ['newDocuments', 'modifiedDocuments', 'skills'], support: ['supports'] }

// Agent applications: associated projects are tasks whose type is "code".
const codeProjects = computed(() => tasks.value.filter(task => task.taskType === 'code'))
function agentLabel(agent: { name:string; model:string }) { return `${agent.name}-${agent.model}` }
function agentById(id: string) { return agents.value.find(agent => agent.id === id) || null }
function agentText(id: string) { const agent = agentById(id); return agent ? agentLabel(agent) : '' }
// Only available Agents are selectable, plus the one already attached, so a task never
// loses its association silently when that Agent goes offline.
function candidateAgents(task: { agentId:string } | null) { if (!task) return agents.value.filter(agent => agent.available); const list = agents.value.filter(agent => agent.available); if (task.agentId && !list.some(agent => agent.id === task.agentId)) { const current = agentById(task.agentId); if (current) return [...list, current] } return list }
function projectNames(agent: AgentApp) { return agent.projectIds.map(id => tasks.value.find(task => task.id === id)?.name || '').filter(name => !!name) }
function resetText(value: string) { return value ? value.replace('T', ' ') : '未设置' }
function setTaskType(target: { taskType:TaskType; agentId:string }, value: TaskType) { target.taskType = value; if (value !== 'code') target.agentId = '' }
function setAgentAvailable(target: { available:boolean; resetAt:string }, value: boolean) { target.available = value; if (value) target.resetAt = '' }
// 快捷冷却：立刻置为不可用，并把重置时间设为现在 + N 小时（分钟精度，与后端 agentResetLayout 一致）。
function setAgentCooldown(target: { available:boolean; resetAt:string }, hours: number) {
  const at = new Date(Date.now() + hours * 3600 * 1000)
  const pad = (value: number) => String(value).padStart(2, '0')
  target.available = false
  target.resetAt = `${at.getFullYear()}-${pad(at.getMonth() + 1)}-${pad(at.getDate())}T${pad(at.getHours())}:${pad(at.getMinutes())}`
}

const parentTasks = computed(() => tasks.value.filter(task => task.kind === 'main' || task.kind === 'branch'))
const filteredTasks = computed(() => tasks.value.filter(task => (status.value === 'all' || task.status === status.value) && (kindFilter.value === 'all' || task.kind === kindFilter.value) && (typeFilter.value === 'all' || task.taskType === typeFilter.value) && `${task.name} ${task.unit} ${task.issuer}`.toLowerCase().includes(query.value.toLowerCase())))
// 列表按「上级任务 → 子任务」分组渲染。子任务进了列表（例如筛「进行中」而上级已完成后被筛掉），
// 它的上级任务即使不满足筛选条件也要显示出来，否则子任务会悬空、看不出归属。
// 这类只为挂靠而出现的上级任务标记为 context，用暗色样式区分，不代表它命中了筛选条件。
const taskGroups = computed(() => {
  const visible = filteredTasks.value
  const visibleIds = new Set(visible.map(task => task.id))
  const groups: { parent: Task; context: boolean; children: Task[] }[] = []
  for (const task of tasks.value) {
    if (task.kind === 'child') continue
    const children = visible.filter(child => child.kind === 'child' && child.parentId === task.id)
    if (!visibleIds.has(task.id) && !children.length) continue
    groups.push({ parent: task, context: !visibleIds.has(task.id), children })
  }
  return groups
})
// 上级任务已被删除（parentId 悬空）的子任务，只能单独列出。
const orphanChildren = computed(() => filteredTasks.value.filter(child => child.kind === 'child' && !tasks.value.some(task => task.id === child.parentId)))
const outputTasks = computed(() => tasks.value.filter(task => task.status === 'completed' && (outputFilter.value === 'all' || task.taskType === outputFilter.value)))
const outputTotals = computed(() => tasks.value.reduce((total, task) => { for (const key of Object.keys(outputLabels) as OutputKey[]) total[key] += task.output[key] || 0; return total }, { newFeatures: 0, modifiedFeatures: 0, applications: 0, skills: 0, newDocuments: 0, modifiedDocuments: 0, supports: 0 } as Record<OutputKey, number>))
const outputCards = computed(() => (outputFilter.value === 'all' ? Object.keys(outputLabels) as OutputKey[] : typeKeys[outputFilter.value]).map(key => ({ key, label: outputLabels[key], value: outputTotals.value[key] })))
const outputMenuItems = computed(() => [{ id: 'all' as const, label: '全部产出', count: Object.values(outputTotals.value).reduce((sum, value) => sum + value, 0) }, ...(['code', 'document', 'support'] as TaskType[]).map(id => ({ id, label: id === 'code' ? '代码' : id === 'document' ? '文档' : '实施', count: tasks.value.filter(task => task.taskType === id).reduce((sum, task) => sum + typeKeys[id].reduce((inner, key) => inner + (task.output[key] || 0), 0), 0) }))])

const dateKey = (date: Date) => `${date.getFullYear()}-${String(date.getMonth() + 1).padStart(2, '0')}-${String(date.getDate()).padStart(2, '0')}`
const monthTitle = computed(() => `${calendarMonth.value.getFullYear()} 年 ${calendarMonth.value.getMonth() + 1} 月`)
const calendarDays = computed(() => { const first = new Date(calendarMonth.value.getFullYear(), calendarMonth.value.getMonth(), 1); const start = new Date(first); start.setDate(1 - first.getDay()); return Array.from({ length: 42 }, (_, index) => { const date = new Date(start); date.setDate(start.getDate() + index); return { date, key: dateKey(date), current: date.getMonth() === calendarMonth.value.getMonth() } }) })
function calendarTasks(key: string) { return filteredTasks.value.filter(task => { if (!task.publishDate) return false; const end = task.status === 'in_progress' ? (task.dueDate && task.dueDate < today() ? task.dueDate : today()) : (task.completedAt || task.dueDate || task.publishDate); return key >= task.publishDate && key <= end }) }
function previousMonth() { calendarMonth.value = new Date(calendarMonth.value.getFullYear(), calendarMonth.value.getMonth() - 1, 1) }
function nextMonth() { calendarMonth.value = new Date(calendarMonth.value.getFullYear(), calendarMonth.value.getMonth() + 1, 1) }

function weekStart(date: string) { const value = new Date(`${date}T00:00:00`); value.setDate(value.getDate() - ((value.getDay() + 6) % 7)); return value.toISOString().slice(0, 10) }
const pieColors = ['#2563eb', '#38bdf8', '#8b5cf6', '#14b8a6', '#f59e0b', '#f43f5e', '#ec4899']
const pieItems = computed(() => { const total = Math.max(1, outputCards.value.reduce((sum, item) => sum + item.value, 0)); let offset = 0; return outputCards.value.map((item, index) => { const start = offset; offset += item.value / total * 100; return { ...item, color: pieColors[index], start, end: offset, percent: Math.round(item.value / total * 100) } }) })
const pieStyle = computed(() => ({ background: `conic-gradient(${pieItems.value.map(item => `${item.color} ${item.start}% ${item.end}%`).join(', ')})` }))

function normalizeTask(value: any): Task { return { ...value, agentId: value.agentId || '', output: { ...emptyOutput(), ...(value.output || {}) } } }
function normalizeAgent(value: any): AgentApp { return { ...value, projectIds: Array.isArray(value?.projectIds) ? value.projectIds.filter((id: any) => !!id) : [], available: !!value?.available, resetAt: value?.resetAt || '' } }
function taskOutputText(task: Task) { return typeKeys[task.taskType].map(key => `${outputLabels[key]} ${task.output[key] || 0}`).join(' | ') }
function setKind(target: { kind: Kind; parentId: string }, kind: Kind) { target.kind = kind; target.parentId = '' }
function setSelf(target: { issuer: string }, event: Event) { target.issuer = (event.target as HTMLInputElement).checked ? '本人' : '' }
function setUnlimited(target: { dueDate: string }, event: Event) { target.dueDate = (event.target as HTMLInputElement).checked ? '' : today() }

function applyBackground() { document.documentElement.style.setProperty('--page-background', settings.backgroundColor) }
async function refresh() { const [taskData, unitData, defaultID, appSettings, agentData] = await Promise.all([ListTasks(), ListUnits(), GetDefaultUnitID(), GetSettings(), ListAgentApps()]); tasks.value = ((taskData ?? []) as any[]).map(normalizeTask); units.value = unitData as Unit[]; agents.value = ((agentData ?? []) as any[]).map(normalizeAgent); defaultUnitID.value = defaultID || ''; Object.assign(settings, appSettings || {}); applyBackground(); selected.value = selected.value ? tasks.value.find(task => task.id === selected.value?.id) || null : tasks.value[0] || null; selectedAgent.value = selectedAgent.value ? agents.value.find(agent => agent.id === selectedAgent.value?.id) || null : agents.value[0] || null; markAllSaved() }
function resetDraft(parent?: Task) { Object.assign(draft, { name: '', kind: parent ? 'child' : 'main', parentId: parent?.id || '', issuer: '', unit: parent?.unit || units.value.find(unit => unit.id === defaultUnitID.value)?.name || units.value[0]?.name || '', publishDate: today(), dueDate: today(), taskType: 'code', agentId: '' }) }
function resetAgentDraft() { Object.assign(agentDraft, { name: '', model: '', projectIds: [], available: true, resetAt: '' }) }
function openAgentCreate() { resetAgentDraft(); showAgentCreate.value = true }
function openCreate() { if (!units.value.length) { section.value = 'units'; message.value = '请先创建一个单位'; return }; resetDraft(); showCreate.value = true }
function openChildCreate(parent: Task) { contextMenu.visible = false; resetDraft(parent); showCreate.value = true }
function showContextMenu(task: Task, event: MouseEvent) { if (task.kind === 'child') return; contextTask.value = task; contextMenu.x = event.clientX; contextMenu.y = event.clientY; contextMenu.visible = true }
function hideContextMenu() { contextMenu.visible = false }
async function create() { try { const created = await CreateTask({ ...draft }) as any; showCreate.value = false; await refresh(); selected.value = tasks.value.find(task => task.id === created.id) || null; resetDraft(); message.value = '任务已创建' } catch (error) { message.value = String(error) } }
async function changeStatus(value: Status) { if (!selected.value) return; await UpdateTaskStatus(selected.value.id, value); await refresh() }
async function createAgent() { try { const created = await CreateAgentApp({ ...agentDraft, projectIds: [...agentDraft.projectIds] }) as any; showAgentCreate.value = false; await refresh(); selectedAgent.value = agents.value.find(agent => agent.id === created.id) || null; resetAgentDraft(); message.value = 'Agent 应用已创建' } catch (error) { message.value = String(error) } }
async function deleteAgent() { const agent = selectedAgent.value; if (!agent) return; requestConfirmation('删除 Agent 应用', `确定删除“${agentLabel(agent)}”吗？删除后无法恢复，相关任务的关联会被清空。`, async () => { try { forgetPending(agent.id); await DeleteAgentApp(agent.id); if (selectedAgent.value?.id === agent.id) selectedAgent.value = null; await refresh(); message.value = 'Agent 应用已删除' } catch (error) { message.value = String(error) } }, '删除') }
async function addUnit() { try { await CreateUnit(newUnit.value); newUnit.value = ''; await refresh() } catch (error) { message.value = String(error) } }
async function rename(unit: Unit) { const name = prompt('修改单位名称', unit.name); if (name !== null) { try { await UpdateUnit(unit.id, name); await refresh() } catch (error) { message.value = String(error) } } }
async function setDefault(id: string) { await SetDefaultUnit(id); await refresh() }
function requestConfirmation(title: string, content: string, action: () => Promise<void>, confirmText = '确认') { Object.assign(confirmation, { visible: true, title, content, confirmText, action }) }
function cancelConfirmation() { confirmation.visible = false; confirmation.action = null }
async function confirmAction() { const action = confirmation.action; cancelConfirmation(); if (action) await action() }
function deleteTask() { const task = selected.value; if (!task) return; requestConfirmation('删除任务', `确定删除任务“${task.name}”吗？删除后无法恢复。`, async () => { try { forgetPending(task.id); await DeleteTask(task.id); if (selected.value?.id === task.id) selected.value = null; await refresh(); message.value = '任务已删除' } catch (error) { message.value = String(error) } }, '删除') }
function deleteUnit(unit: Unit) { requestConfirmation('删除单位', `确定删除单位“${unit.name}”吗？删除后无法恢复。`, async () => { try { await DeleteUnit(unit.id); await refresh(); message.value = '单位已删除' } catch (error) { message.value = String(error) } }, '删除') }
async function saveSettings() { try { await UpdateSettings({ ...settings }); applyBackground(); showSettings.value = false; message.value = '设置已保存' } catch (error) { message.value = String(error) } }
// ---- 即时保存 ----
// 只把面板可编辑的字段纳入签名：createdAt / updatedAt 由服务端每次保存后刷新，
// 一旦纳入就会造成「保存 → 回写 → 签名变化 → 再保存」的死循环。
function taskPayload(task: Task) { return { id: task.id, name: task.name, kind: task.kind, parentId: task.parentId, issuer: task.issuer, unit: task.unit, publishDate: task.publishDate, dueDate: task.dueDate, taskType: task.taskType, status: task.status, completedAt: task.completedAt, agentId: task.agentId, output: { ...task.output } } }
function agentPayload(agent: AgentApp) { return { name: agent.name, model: agent.model, projectIds: [...agent.projectIds], available: agent.available, resetAt: agent.resetAt } }
function signature(payload: unknown) { return JSON.stringify(payload) }
function markSaved(id: string, payload: unknown) { if (dirtyTasks.has(id) || dirtyAgents.has(id)) return; savedSignatures.set(id, signature(payload)) }
function markAllSaved() { tasks.value.forEach(task => savedSignatures.set(task.id, signature(taskPayload(task)))); agents.value.forEach(agent => savedSignatures.set(agent.id, signature(agentPayload(agent)))) }
function forgetPending(id: string) { dirtyTasks.delete(id); dirtyAgents.delete(id); savedSignatures.delete(id) }

// 落盘前先在本地拦一道，清空输入框准备重打时不会反复弹错误提示。
function taskIssue(task: Task) { if (!task.name.trim() || !task.issuer.trim() || !task.unit.trim()) return '任务名、下发人和单位不能为空'; if (!task.publishDate) return '请选择发布日期'; if (task.dueDate && task.dueDate < task.publishDate) return '截止日期不能早于发布日期'; if (task.kind === 'child' && !task.parentId) return '请选择所属主线任务'; return '' }
function agentIssue(agent: AgentApp) { return agent.name.trim() && agent.model.trim() ? '' : '应用名和模型名不能为空' }

function queueTask(task: Task) { if (signature(taskPayload(task)) === savedSignatures.get(task.id)) return; dirtyTasks.set(task.id, task); scheduleSave() }
function queueAgent(agent: AgentApp) { if (signature(agentPayload(agent)) === savedSignatures.get(agent.id)) return; dirtyAgents.set(agent.id, agent); scheduleSave() }
function scheduleSave() { if (saveTimer) clearTimeout(saveTimer); saveTimer = setTimeout(() => { void flushSaves() }, SAVE_DEBOUNCE) }

// 任务侧是唯一事实来源，Agent 的项目列表只是它的视图，这里做同样的本地镜像，
// 免得为了显示一致再跑一趟后端。
function mirrorAgentProjects() { for (const agent of agents.value) { const ids = tasks.value.filter(task => task.taskType === 'code' && task.agentId === agent.id).map(task => task.id); if (ids.join('|') === agent.projectIds.join('|')) continue; agent.projectIds = ids; markSaved(agent.id, agentPayload(agent)) } }
function mirrorTaskAgents(agent: AgentApp) { for (const task of tasks.value) { if (task.taskType !== 'code') continue; const next = agent.projectIds.includes(task.id) ? agent.id : (task.agentId === agent.id ? '' : task.agentId); if (next === task.agentId) continue; task.agentId = next; markSaved(task.id, taskPayload(task)) } }

async function flushSaves() {
  if (saveTimer) { clearTimeout(saveTimer); saveTimer = undefined }
  const queuedTasks = [...dirtyTasks.values()]; dirtyTasks.clear()
  const queuedAgents = [...dirtyAgents.values()]; dirtyAgents.clear()
  if (!queuedTasks.length && !queuedAgents.length) return
  try {
    for (const task of queuedTasks) {
      // 校验未过的改动不下发，也不打断其它字段的保存，只弹出一次提示。
      const issue = taskIssue(task)
      if (issue) { message.value = issue; continue }
      const saved = normalizeTask(await UpdateTask(taskPayload(task) as any))
      savedSignatures.set(saved.id, signature(taskPayload(saved)))
      Object.assign(task, saved)
      const row = tasks.value.find(item => item.id === saved.id)
      if (row && row !== task) Object.assign(row, saved)
      mirrorAgentProjects()
    }
    for (const agent of queuedAgents) {
      const issue = agentIssue(agent)
      if (issue) { message.value = issue; continue }
      const saved = normalizeAgent(await UpdateAgentApp(agent.id, agentPayload(agent) as any))
      // 保留对象引用：替换它会让正在编辑的表单失去绑定。
      Object.assign(agent, saved)
      savedSignatures.set(agent.id, signature(agentPayload(agent)))
      mirrorTaskAgents(agent)
    }
  } catch (error) {
    message.value = String(error)
  }
}

function resetBackground() { settings.backgroundColor = '#f7faff'; applyBackground() }
let toastTimer: ReturnType<typeof setTimeout> | undefined
watch(message, value => { if (toastTimer) clearTimeout(toastTimer); if (value) toastTimer = setTimeout(() => { message.value = '' }, 3000) })
// 面板里的改动先入队，防抖后统一落盘；签名一致时 queue* 会直接跳过，
// 因此「切换选中项」与「服务端回写」都不会触发多余的保存。
watch(selected, value => { if (value) queueTask(value) }, { deep: true })
watch(selectedAgent, value => { if (value) queueAgent(value) }, { deep: true })
onMounted(() => {
  refresh().catch(error => message.value = String(error))
  // 重置时间到点后由后端广播，这里刷新列表即可看到最新状态。
  try { EventsOn('agents:reset', () => { message.value = 'Agent 重置时间已到，已恢复可用'; refresh().catch(() => {}) }) } catch { /* 无 Wails 运行时的预览环境没有事件通道 */ }
})
</script>

<template>
  <main class="shell" @click="hideContextMenu">
    <aside class="sidebar">
      <div class="brand"><img class="brand-mark" :src="appLogo" alt="Task Manager"><div><b>Task Manager</b><small>工作任务管理器</small></div></div>
      <p class="side-title">任务概览</p>
      <button v-for="item in [{ id: 'all', label: '全部任务' }, { id: 'in_progress', label: '进行中' }, { id: 'completed', label: '已完成' }, { id: 'cancelled', label: '已取消' }]" :key="item.id" class="nav-item" :class="{ active: section === 'tasks' && status === item.id }" @click="section = 'tasks'; status = item.id as 'all' | Status"><span>{{ item.label }}</span><em>{{ item.id === 'all' ? tasks.length : tasks.filter(task => task.status === item.id).length }}</em></button>
      <p class="side-title menu-divider">产出概览</p>
      <button v-for="item in outputMenuItems" :key="item.id" class="nav-item" :class="{ active: section === 'output' && outputFilter === item.id }" @click="section = 'output'; outputFilter = item.id"><span>{{ item.label }}</span><em>{{ item.count }}</em></button>
      <p class="side-title menu-divider">Agent</p>
      <button class="nav-item nav-sub" :class="{ active: section === 'agent' }" @click="section = 'agent'"><span>Agent应用</span><em>{{ agents.length }}</em></button>
      <button class="nav-item menu-divider" :class="{ active: section === 'units' }" @click="section = 'units'"><span>单位管理</span><em>{{ units.length }}</em></button>
      <button class="nav-item settings-button" @click="showSettings = true"><span>⚙ 设置</span></button>
      <button class="nav-item about-button" @click="showAbout = true"><span>ⓘ 关于</span></button>
    </aside>

    <section class="workspace">
      <template v-if="section === 'tasks'">
        <header><div><p class="eyebrow">工作空间</p><div class="title-line"><h1>我的任务</h1><div class="view-switch"><button :class="{ active: viewMode === 'list' }" @click="viewMode = 'list'">列表</button><button :class="{ active: viewMode === 'calendar' }" @click="viewMode = 'calendar'">日历</button></div></div></div><div class="actions"><button @click="ImportData().then(refresh)">导入 JSON</button><button @click="ExportData">导出 JSON</button><button class="primary" @click="openCreate">＋ 新建任务</button></div></header>
        <div class="toolbar"><input v-model="query" placeholder="搜索任务名、下发人或单位"><select v-model="kindFilter"><option value="all">全部层级</option><option v-for="(_, key) in kindText" :key="key" :value="key">{{ kindText[key as Kind] }}</option></select><select v-model="typeFilter"><option value="all">全部类型</option><option value="code">代码</option><option value="document">文档</option><option value="support">实施</option></select></div>
        <div v-if="viewMode === 'list'" class="task-list"><template v-for="group in taskGroups" :key="group.parent.id"><button class="task-row" :class="{ completed: group.parent.status === 'completed', context: group.context }" @click="selected = group.parent" @contextmenu.prevent="showContextMenu(group.parent, $event)"><span>{{ group.parent.kind === 'main' ? '◆' : '◇' }}</span><span class="task-name">{{ group.parent.name }}<small>{{ kindText[group.parent.kind] }} · {{ group.parent.unit }} · {{ group.parent.issuer }}{{ group.parent.agentId ? ' · ' + agentText(group.parent.agentId) : '' }}</small></span><span class="date">{{ group.parent.status === 'completed' ? group.parent.completedAt : group.parent.dueDate || '无期限' }}</span><span class="badge" :class="group.parent.status">{{ statusText[group.parent.status] }}</span></button><button v-for="child in group.children" :key="child.id" class="task-row child" :class="{ completed: child.status === 'completed' }" @click="selected = child"><span>↳</span><span class="task-name">{{ child.name }}<small>子任务 · {{ child.unit }} · {{ child.issuer }}{{ child.agentId ? ' · ' + agentText(child.agentId) : '' }}</small></span><span class="date">{{ child.status === 'completed' ? child.completedAt : child.dueDate || '无期限' }}</span><span class="badge" :class="child.status">{{ statusText[child.status] }}</span></button></template><button v-for="child in orphanChildren" :key="child.id" class="task-row child orphan-child" :class="{ completed: child.status === 'completed' }" @click="selected = child"><span>↳</span><span class="task-name">{{ child.name }}<small>子任务（上级已删除） · {{ child.unit }} · {{ child.issuer }}{{ child.agentId ? ' · ' + agentText(child.agentId) : '' }}</small></span><span class="date">{{ child.status === 'completed' ? child.completedAt : child.dueDate || '无期限' }}</span><span class="badge" :class="child.status">{{ statusText[child.status] }}</span></button><div v-if="!taskGroups.length && !orphanChildren.length" class="empty">暂无任务。</div></div>
        <div v-else class="calendar-view"><div class="calendar-head"><button @click="previousMonth">‹</button><strong>{{ monthTitle }}</strong><button @click="nextMonth">›</button></div><div class="calendar-scroll"><div class="calendar-week"><span v-for="day in ['日','一','二','三','四','五','六']" :key="day">周{{ day }}</span></div><div class="calendar-grid"><div v-for="day in calendarDays" :key="day.key" class="calendar-cell" :class="{ muted: !day.current, today: day.key === today() }"><b>{{ day.date.getDate() }}</b><button v-for="task in calendarTasks(day.key)" :key="task.id" class="calendar-task" :class="task.status" @click="selected = task">{{ task.name }}</button></div></div></div></div>
      </template>

      <div v-else-if="section === 'output'" class="output-dashboard">
        <header><div><p class="eyebrow">产出概览</p><h1>{{ outputFilter === 'all' ? '全部产出' : outputFilter === 'code' ? '代码产出' : outputFilter === 'document' ? '文档产出' : '实施产出' }}</h1></div></header>
        <div class="output-cards"><article v-for="card in outputCards" :key="card.key" class="output-card"><span>{{ card.label }}</span><strong>{{ card.value }}</strong><small>累计数量</small></article></div>
        <section v-if="outputFilter === 'all'" class="chart-panel"><h2>产出项目占比</h2><div class="pie-layout"><div class="pie-chart" :style="pieStyle"></div><div class="pie-legend"><div v-for="item in pieItems" :key="item.key"><i :style="{ backgroundColor: item.color }"></i><span>{{ item.label }}</span><b>{{ item.value }}（{{ item.percent }}%）</b></div></div></div></section>
        <section v-else class="output-list"><h2>对应任务</h2><button v-for="task in outputTasks" :key="task.id" class="output-task" @click="selected = task; section = 'tasks'"><span><b>{{ task.name }}</b><small>{{ task.unit }} · {{ task.issuer }} · {{ statusText[task.status] }}</small></span><em>{{ taskOutputText(task) }}</em></button><div v-if="!outputTasks.length" class="empty">暂无对应任务。</div></section>
      </div>
      <div v-else-if="section === 'agent'" class="agents-workspace">
        <header><div><p class="eyebrow">Agent 管理</p><h1>Agent应用</h1></div><div class="actions"><button class="primary" @click="openAgentCreate">＋ 新增 Agent 应用</button></div></header>
        <section class="agents-panel"><div class="agent-list"><button v-for="agent in agents" :key="agent.id" class="agent-row" :class="{ selected: selectedAgent?.id === agent.id }" @click="selectedAgent = agent"><span class="agent-title">{{ agentLabel(agent) }}</span><span class="agent-projects">{{ projectNames(agent).length ? projectNames(agent).join('、') : '未关联项目' }}</span><span class="agent-reset">{{ agent.available ? '—' : resetText(agent.resetAt) }}</span><span class="agent-pill" :class="{ offline: !agent.available }">{{ agent.available ? '可用' : '不可用' }}</span></button><div v-if="!agents.length" class="empty">暂无 Agent 应用，点击右上角新增。</div></div></section>
      </div>
      <div v-else class="units-workspace"><header><div><p class="eyebrow">基础信息</p><h1>单位管理</h1></div></header><section class="units-panel"><form class="unit-add" @submit.prevent="addUnit"><input v-model="newUnit" placeholder="输入单位名称"><button class="primary">新增单位</button></form><div class="unit-list"><div v-for="unit in units" :key="unit.id" class="unit-row"><span>{{ unit.name }}</span><em v-if="unit.id === defaultUnitID">默认单位</em><button @click="setDefault(unit.id)">设为默认</button><button @click="rename(unit)">修改</button><button class="delete-button" @click="deleteUnit(unit)">删除</button></div><div v-if="!units.length" class="empty">暂无单位，请先新增一个单位。</div></div></section></div>
    </section>

    <aside v-if="section !== 'agent' && selected" class="details"><div class="details-heading"><div><p class="eyebrow">任务详情</p><input v-model="selected.name"></div><div class="heading-status"><span class="badge" :class="selected.status">{{ statusText[selected.status] }}</span></div></div><div class="status-buttons"><button v-for="item in ['in_progress','completed','cancelled'] as Status[]" :key="item" :class="{ chosen: selected.status === item }" @click="changeStatus(item)">{{ statusText[item] }}</button></div><div class="fields"><label>单位<select v-model="selected.unit"><option v-for="unit in units" :key="unit.id" :value="unit.name">{{ unit.name }}</option></select></label><label>任务类型<select :value="selected.taskType" @change="setTaskType(selected!, ($event.target as HTMLSelectElement).value as TaskType)"><option value="code">代码</option><option value="document">文档</option><option value="support">实施</option></select></label><label v-if="selected.taskType === 'code'" class="wide-field">关联 Agent<select v-model="selected.agentId"><option value="">不关联</option><option v-for="agent in candidateAgents(selected)" :key="agent.id" :value="agent.id">{{ agentLabel(agent) }}{{ agent.available ? '' : '（不可用）' }}</option></select></label><label>层级<select :value="selected.kind" @change="setKind(selected!, ($event.target as HTMLSelectElement).value as Kind)"><option v-for="(_, key) in kindText" :key="key" :value="key">{{ kindText[key as Kind] }}</option></select></label><label v-if="selected.kind === 'child'">所属任务<select v-model="selected.parentId"><option value="">请选择</option><option v-for="task in parentTasks" :key="task.id" :value="task.id">{{ task.name }}（{{ kindText[task.kind] }}）</option></select></label><label>下发人<input v-model="selected.issuer" :disabled="selected.issuer === '本人'"><span class="inline-toggle"><input type="checkbox" :checked="selected.issuer === '本人'" @change="setSelf(selected!, $event)"><b>本人</b></span></label><label>发布日期<input v-model="selected.publishDate" type="date"></label><label>截止日期<input v-if="selected.dueDate" v-model="selected.dueDate" type="date"><span v-else class="empty-date-field"></span><span class="inline-toggle"><input type="checkbox" :checked="!selected.dueDate" @change="setUnlimited(selected!, $event)"><b>无期限</b></span></label><label v-if="selected.status === 'completed'">完成日期<input v-model="selected.completedAt" type="date"></label></div><section class="output-editor"><h3>产出</h3><label>结果<textarea v-model="selected.output.result"></textarea></label><div v-if="selected.taskType !== 'support'" class="quantity-grid"><template v-for="key in typeKeys[selected.taskType]" :key="key"><label>{{ outputLabels[key] }}<OutputStepper :model-value="selected.output[key]" @update:model-value="selected.output[key] = $event"/></label></template></div><p v-else class="support-hint">实施任务完成后自动计为 1 项实施。</p></section><div class="detail-actions"><button class="delete-button" @click="deleteTask">删除任务</button></div></aside>
    <aside v-else-if="selectedAgent" class="details"><div class="details-heading"><div><p class="eyebrow">Agent 应用</p><span class="agent-heading">{{ agentLabel(selectedAgent) }}</span></div><div class="heading-status"><span class="agent-pill" :class="{ offline: !selectedAgent.available }">{{ selectedAgent.available ? '可用' : '不可用' }}</span></div></div><div class="fields"><label>应用名<input v-model="selectedAgent.name"></label><label>模型名<input v-model="selectedAgent.model" placeholder="例如 Claude Opus 4.1"></label></div><label>可用情况<span class="inline-toggle"><input type="checkbox" :checked="selectedAgent.available" @change="setAgentAvailable(selectedAgent!, ($event.target as HTMLInputElement).checked)"><b>当前可用</b></span></label><div class="cooldown-row"><button type="button" class="cooldown-button" @click="setAgentCooldown(selectedAgent!, 5)">5小时后可用</button><small class="hint">点击后立即置为不可用，5 小时后自动恢复并发送通知</small></div><label v-if="!selectedAgent.available">下次重置时间<input v-model="selectedAgent.resetAt" type="datetime-local"></label><section class="agent-projects-editor"><h3>关联项目</h3><p class="hint">仅可选择任务类型为「代码」的任务</p><div class="agent-project-list"><label v-for="task in codeProjects" :key="selectedAgent.id + ':' + task.id" class="agent-project-item"><input type="checkbox" :value="task.id" v-model="selectedAgent.projectIds"><span>{{ task.name }}</span></label><p v-if="!codeProjects.length" class="hint">暂无可选的代码任务。</p></div></section><div class="detail-actions"><button class="delete-button" @click="deleteAgent">删除应用</button></div></aside>
    <aside v-else class="details empty">{{ section === 'agent' ? '选择一项 Agent 应用' : '选择一项任务' }}</aside>

    <div v-if="showCreate" class="modal-backdrop"><form class="modal" @submit.prevent="create"><div class="modal-title"><h2>创建任务</h2><button class="close" type="button" @click="showCreate = false">×</button></div><div class="form-grid"><label>单位<select v-model="draft.unit" required><option v-for="unit in units" :key="unit.id" :value="unit.name">{{ unit.name }}</option></select></label><label>任务层级<select :value="draft.kind" @change="setKind(draft, ($event.target as HTMLSelectElement).value as Kind)"><option v-for="(_, key) in kindText" :key="key" :value="key">{{ kindText[key as Kind] }}</option></select></label><label>任务名<input v-model="draft.name" required></label><label>下发人<input v-model="draft.issuer" required :disabled="draft.issuer === '本人'"><span class="inline-toggle"><input type="checkbox" :checked="draft.issuer === '本人'" @change="setSelf(draft, $event)"><b>本人</b></span></label><label v-if="draft.kind === 'child'">所属任务<select v-model="draft.parentId" required><option value="">请选择</option><option v-for="task in parentTasks" :key="task.id" :value="task.id">{{ task.name }}</option></select></label><label>发布日期<input v-model="draft.publishDate" type="date" required></label><label>截止日期<input v-if="draft.dueDate" v-model="draft.dueDate" type="date"><span v-else class="empty-date-field"></span><span class="inline-toggle"><input type="checkbox" :checked="!draft.dueDate" @change="setUnlimited(draft, $event)"><b>无期限</b></span></label><label>任务类型<select :value="draft.taskType" @change="setTaskType(draft, ($event.target as HTMLSelectElement).value as TaskType)"><option value="code">代码</option><option value="document">文档</option><option value="support">实施</option></select></label><label v-if="draft.taskType === 'code'" class="wide-field">关联 Agent<select v-model="draft.agentId"><option value="">不关联</option><option v-for="agent in candidateAgents(draft)" :key="agent.id" :value="agent.id">{{ agentLabel(agent) }}{{ agent.available ? '' : '（不可用）' }}</option></select></label></div><div class="actions"><button type="button" @click="showCreate = false">取消</button><button class="primary">创建任务</button></div></form></div>
    <div v-if="showAgentCreate" class="modal-backdrop"><form class="modal agent-modal" @submit.prevent="createAgent"><div class="modal-title"><h2>新增 Agent 应用</h2><button class="close" type="button" @click="showAgentCreate = false">×</button></div><div class="form-grid"><label>应用名<input v-model="agentDraft.name" placeholder="例如 Claude Code" required></label><label>模型名<input v-model="agentDraft.model" placeholder="例如 Claude Opus 4.1" required></label></div><label class="setting-row"><span><b>可用情况</b><small>当前该模型是否可用</small></span><input v-model="agentDraft.available" type="checkbox" class="setting-toggle"></label><label v-if="!agentDraft.available" class="setting-row"><span><b>下次重置时间</b><small>精确到分钟，到点后自动恢复可用</small></span><input v-model="agentDraft.resetAt" type="datetime-local"></label><section class="agent-projects-editor"><h3>关联项目</h3><p class="hint">仅可选择任务类型为「代码」的任务</p><div class="agent-project-list"><label v-for="task in codeProjects" :key="task.id" class="agent-project-item"><input type="checkbox" :value="task.id" v-model="agentDraft.projectIds"><span>{{ task.name }}</span></label><p v-if="!codeProjects.length" class="hint">暂无可选的代码任务。</p></div></section><div class="actions"><button type="button" @click="showAgentCreate = false">取消</button><button class="primary">创建 Agent 应用</button></div></form></div>
    <div v-if="confirmation.visible" class="modal-backdrop confirmation-backdrop" @click.self="cancelConfirmation"><section class="modal confirmation-modal" role="dialog" aria-modal="true" :aria-label="confirmation.title"><div class="confirmation-icon">!</div><h2>{{ confirmation.title }}</h2><p>{{ confirmation.content }}</p><div class="actions"><button type="button" @click="cancelConfirmation">取消</button><button type="button" class="primary danger-confirm" @click="confirmAction">{{ confirmation.confirmText }}</button></div></section></div>
    <div v-if="contextMenu.visible && contextTask" class="task-context-menu" :style="{ left: `${contextMenu.x}px`, top: `${contextMenu.y}px` }" @click.stop><button @click="openChildCreate(contextTask)">添加子任务</button></div>
    <button v-if="message" class="toast" @click="message = ''">{{ message }}</button>
    <div v-if="showSettings" class="modal-backdrop"><form class="modal settings-modal" @submit.prevent="saveSettings"><div class="modal-title"><h2>设置</h2><button class="close" type="button" @click="showSettings = false">×</button></div><label class="setting-row"><span><b>主页面背景色</b><small>影响任务与产出主页面背景</small></span><input v-model="settings.backgroundColor" type="color" @input="applyBackground"></label><div class="preset-colors"><button v-for="color in ['#f7faff','#f1f5f9','#f8fafc','#eef6ff','#f5f3ff','#fffaf0']" :key="color" type="button" :style="{ backgroundColor: color }" :class="{ selected: settings.backgroundColor === color }" @click="settings.backgroundColor = color; applyBackground()"></button><button type="button" class="reset-color" @click="resetBackground">恢复默认</button></div><label class="setting-row"><span><b>开机自动启动</b><small>登录 Windows 后自动打开 Task Manager</small></span><input v-model="settings.autoStart" type="checkbox" class="setting-toggle"></label><div class="actions"><button type="button" @click="showSettings = false">取消</button><button class="primary">保存设置</button></div></form></div>
    <div v-if="showAbout" class="modal-backdrop"><div class="modal about-modal"><div class="modal-title"><h2>关于</h2><button class="close" @click="showAbout = false">×</button></div><div class="about-grid"><div><small>程序名</small><strong>Task Manager | 工作任务管理器</strong></div><div><small>版本号</small><strong>beta</strong></div><div><small>作者</small><strong>李正浩</strong></div></div></div></div>
  </main>
</template>



