<script setup lang="ts">
import { computed, defineComponent, h, onMounted, reactive, ref } from 'vue'
import { CreateTask, CreateUnit, ExportData, GetDefaultUnitID, ImportData, ListTasks, ListUnits, SetDefaultUnit, UpdateTask, UpdateTaskStatus, UpdateUnit } from '../wailsjs/go/main/App'

type Status = 'in_progress' | 'completed' | 'cancelled'
type Kind = 'main' | 'branch' | 'child'
type TaskType = 'code' | 'document' | 'support'
type OutputKey = 'newFeatures' | 'modifiedFeatures' | 'applications' | 'newDocuments' | 'modifiedDocuments' | 'supports'
type Output = { result: string } & Record<OutputKey, number>
type Task = { id:string; name:string; kind:Kind; parentId:string; issuer:string; unit:string; publishDate:string; dueDate:string; taskType:TaskType; status:Status; completedAt:string; createdAt:string; updatedAt:string; output:Output }
type Unit = { id:string; name:string }

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
const selected = ref<Task | null>(null)
const defaultUnitID = ref('')
const section = ref<'tasks' | 'output'>('tasks')
const outputFilter = ref<'all' | TaskType>('all')
const showCreate = ref(false)
const showUnits = ref(false)
const showAbout = ref(false)
const message = ref('')
const newUnit = ref('')
const query = ref('')
const status = ref<'all' | Status>('all')
const kindFilter = ref<'all' | Kind>('all')
const typeFilter = ref<'all' | TaskType>('all')

const today = () => new Date().toISOString().slice(0, 10)
const emptyOutput = (): Output => ({ result: '', newFeatures: 0, modifiedFeatures: 0, applications: 0, newDocuments: 0, modifiedDocuments: 0, supports: 0 })
const draft = reactive({ name: '', kind: 'main' as Kind, parentId: '', issuer: '', unit: '', publishDate: today(), dueDate: today(), taskType: 'code' as TaskType })
const statusText: Record<Status, string> = { in_progress: '进行中', completed: '已完成', cancelled: '已取消' }
const kindText: Record<Kind, string> = { main: '主线任务', branch: '支线任务', child: '子任务' }
const outputLabels: Record<OutputKey, string> = { newFeatures: '新增功能', modifiedFeatures: '修改功能', applications: '应用程序', newDocuments: '新增文档', modifiedDocuments: '修改文档', supports: '支持' }
const typeKeys: Record<TaskType, OutputKey[]> = { code: ['newFeatures', 'modifiedFeatures', 'applications'], document: ['newDocuments', 'modifiedDocuments'], support: ['supports'] }

const mains = computed(() => tasks.value.filter(task => task.kind === 'main'))
const filteredTasks = computed(() => tasks.value.filter(task => (status.value === 'all' || task.status === status.value) && (kindFilter.value === 'all' || task.kind === kindFilter.value) && (typeFilter.value === 'all' || task.taskType === typeFilter.value) && `${task.name} ${task.unit} ${task.issuer}`.toLowerCase().includes(query.value.toLowerCase())))
const outputTasks = computed(() => outputFilter.value === 'all' ? tasks.value : tasks.value.filter(task => task.taskType === outputFilter.value))
const outputTotals = computed(() => tasks.value.reduce((total, task) => { for (const key of Object.keys(outputLabels) as OutputKey[]) total[key] += task.output[key] || 0; return total }, { newFeatures: 0, modifiedFeatures: 0, applications: 0, newDocuments: 0, modifiedDocuments: 0, supports: 0 } as Record<OutputKey, number>))
const outputCards = computed(() => (outputFilter.value === 'all' ? Object.keys(outputLabels) as OutputKey[] : typeKeys[outputFilter.value]).map(key => ({ key, label: outputLabels[key], value: outputTotals.value[key] })))
const outputMenuItems = computed(() => [{ id: 'all' as const, label: '全部产出', count: Object.values(outputTotals.value).reduce((sum, value) => sum + value, 0) }, ...(['code', 'document', 'support'] as TaskType[]).map(id => ({ id, label: id === 'code' ? '代码' : id === 'document' ? '文档' : '支持', count: tasks.value.filter(task => task.taskType === id).reduce((sum, task) => sum + typeKeys[id].reduce((inner, key) => inner + (task.output[key] || 0), 0), 0) }))])

function weekStart(date: string) { const value = new Date(`${date}T00:00:00`); value.setDate(value.getDate() - ((value.getDay() + 6) % 7)); return value.toISOString().slice(0, 10) }
const weeklyTrend = computed(() => { const map = new Map<string, number>(); tasks.value.forEach(task => { const date = task.completedAt || task.publishDate; if (!date) return; const week = weekStart(date); const amount = typeKeys[task.taskType].reduce((sum, key) => sum + (task.output[key] || 0), 0); map.set(week, (map.get(week) || 0) + amount) }); const rows = [...map.entries()].sort(([a], [b]) => a.localeCompare(b)).slice(-8).map(([week, value]) => ({ week, value })); return rows.length ? rows : [{ week: weekStart(today()), value: 0 }] })
const trendMax = computed(() => Math.max(1, ...weeklyTrend.value.map(row => row.value)))
function weekRange(week: string) { const start = new Date(`${week}T00:00:00`); const end = new Date(start); end.setDate(end.getDate() + 6); const format = (date: Date) => `${date.getFullYear()}/${String(date.getMonth() + 1).padStart(2, '0')}/${String(date.getDate()).padStart(2, '0')}`; return `${format(start)}-${format(end)}` }

function normalizeTask(value: any): Task { return { ...value, output: { ...emptyOutput(), ...(value.output || {}) } } }
function taskOutputText(task: Task) { return typeKeys[task.taskType].map(key => `${outputLabels[key]} ${task.output[key] || 0}`).join(' | ') }
function setKind(target: { kind: Kind; parentId: string }, kind: Kind) { target.kind = kind; target.parentId = '' }
function setSelf(target: { issuer: string }, event: Event) { target.issuer = (event.target as HTMLInputElement).checked ? '本人' : '' }
function setUnlimited(target: { dueDate: string }, event: Event) { target.dueDate = (event.target as HTMLInputElement).checked ? '' : today() }

async function refresh() { const [taskData, unitData, defaultID] = await Promise.all([ListTasks(), ListUnits(), GetDefaultUnitID()]); tasks.value = ((taskData ?? []) as any[]).map(normalizeTask); units.value = unitData as Unit[]; defaultUnitID.value = defaultID || ''; selected.value = selected.value ? tasks.value.find(task => task.id === selected.value?.id) || null : tasks.value[0] || null }
function openCreate() { if (!units.value.length) { showUnits.value = true; message.value = '请先创建一个单位'; return }; draft.unit = units.value.find(unit => unit.id === defaultUnitID.value)?.name || units.value[0].name; showCreate.value = true }
async function create() { try { const created = await CreateTask({ ...draft }) as any; showCreate.value = false; await refresh(); selected.value = tasks.value.find(task => task.id === created.id) || null; message.value = '任务已创建' } catch (error) { message.value = String(error) } }
async function save() { if (!selected.value) return; try { await UpdateTask(selected.value as any); await refresh(); message.value = '任务已保存' } catch (error) { message.value = String(error) } }
async function changeStatus(value: Status) { if (!selected.value) return; await UpdateTaskStatus(selected.value.id, value); await refresh() }
async function addUnit() { try { await CreateUnit(newUnit.value); newUnit.value = ''; await refresh() } catch (error) { message.value = String(error) } }
async function rename(unit: Unit) { const name = prompt('修改单位名称', unit.name); if (name !== null) { try { await UpdateUnit(unit.id, name); await refresh() } catch (error) { message.value = String(error) } } }
async function setDefault(id: string) { await SetDefaultUnit(id); await refresh() }
onMounted(() => refresh().catch(error => message.value = String(error)))
</script>

<template>
  <main class="shell">
    <aside class="sidebar">
      <div class="brand"><span class="brand-mark">✓</span><div><b>Task Manager</b><small>工作任务管理器</small></div></div>
      <p class="side-title">任务概览</p>
      <button v-for="item in [{ id: 'all', label: '全部任务' }, { id: 'in_progress', label: '进行中' }, { id: 'completed', label: '已完成' }, { id: 'cancelled', label: '已取消' }]" :key="item.id" class="nav-item" :class="{ active: section === 'tasks' && status === item.id }" @click="section = 'tasks'; status = item.id as 'all' | Status"><span>{{ item.label }}</span><em>{{ item.id === 'all' ? tasks.length : tasks.filter(task => task.status === item.id).length }}</em></button>
      <p class="side-title menu-divider">产出概览</p>
      <button v-for="item in outputMenuItems" :key="item.id" class="nav-item" :class="{ active: section === 'output' && outputFilter === item.id }" @click="section = 'output'; outputFilter = item.id"><span>{{ item.label }}</span><em>{{ item.count }}</em></button>
      <button class="nav-item menu-divider" @click="showUnits = true"><span>单位管理</span><em>{{ units.length }}</em></button>
      <button class="nav-item about-button" @click="showAbout = true"><span>ⓘ 关于</span></button>
    </aside>

    <section class="workspace">
      <template v-if="section === 'tasks'">
        <header><div><p class="eyebrow">工作空间</p><h1>我的任务</h1></div><div class="actions"><button @click="ImportData().then(refresh)">导入 JSON</button><button @click="ExportData">导出 JSON</button><button class="primary" @click="openCreate">＋ 新建任务</button></div></header>
        <div class="toolbar"><input v-model="query" placeholder="搜索任务名、下发人或单位"><select v-model="kindFilter"><option value="all">全部层级</option><option v-for="(_, key) in kindText" :key="key" :value="key">{{ kindText[key as Kind] }}</option></select><select v-model="typeFilter"><option value="all">全部类型</option><option value="code">代码</option><option value="document">文档</option><option value="support">支持</option></select></div>
        <div class="task-list"><template v-for="task in filteredTasks.filter(task => task.kind !== 'child')" :key="task.id"><button class="task-row" :class="{ completed: task.status === 'completed' }" @click="selected = task"><span>{{ task.kind === 'main' ? '◆' : '◇' }}</span><span class="task-name">{{ task.name }}<small>{{ kindText[task.kind] }} · {{ task.unit }} · {{ task.issuer }}</small></span><span class="date">{{ task.status === 'completed' ? task.completedAt : task.dueDate || '无期限' }}</span><span class="badge" :class="task.status">{{ statusText[task.status] }}</span></button><button v-for="child in filteredTasks.filter(child => child.kind === 'child' && child.parentId === task.id)" :key="child.id" class="task-row child" :class="{ completed: child.status === 'completed' }" @click="selected = child"><span>↳</span><span class="task-name">{{ child.name }}<small>子任务 · {{ child.unit }} · {{ child.issuer }}</small></span><span class="date">{{ child.status === 'completed' ? child.completedAt : child.dueDate || '无期限' }}</span><span class="badge" :class="child.status">{{ statusText[child.status] }}</span></button></template><div v-if="!filteredTasks.length" class="empty">暂无任务。</div></div>
      </template>

      <div v-else class="output-dashboard">
        <header><div><p class="eyebrow">产出概览</p><h1>{{ outputFilter === 'all' ? '全部产出' : outputFilter === 'code' ? '代码产出' : outputFilter === 'document' ? '文档产出' : '支持产出' }}</h1></div></header>
        <div class="output-cards"><article v-for="card in outputCards" :key="card.key" class="output-card"><span>{{ card.label }}</span><strong>{{ card.value }}</strong><small>累计数量</small></article></div>
        <section v-if="outputFilter === 'all'" class="chart-panel"><h2>产出趋势</h2><div class="bar-chart"><div v-for="row in weeklyTrend" :key="row.week" class="bar-column"><b>{{ row.value }}</b><i :style="{ height: `${Math.max(5, row.value / trendMax * 100)}%` }"></i><span>{{ weekRange(row.week) }}</span></div></div></section>
        <section v-else class="output-list"><h2>对应任务</h2><button v-for="task in outputTasks" :key="task.id" class="output-task" @click="selected = task; section = 'tasks'"><span><b>{{ task.name }}</b><small>{{ task.unit }} · {{ task.issuer }} · {{ statusText[task.status] }}</small></span><em>{{ taskOutputText(task) }}</em></button><div v-if="!outputTasks.length" class="empty">暂无对应任务。</div></section>
      </div>
    </section>

    <aside v-if="selected" class="details"><div class="details-heading"><div><p class="eyebrow">任务详情（可编辑）</p><input v-model="selected.name"></div><span class="badge" :class="selected.status">{{ statusText[selected.status] }}</span></div><div class="status-buttons"><button v-for="item in ['in_progress','completed','cancelled'] as Status[]" :key="item" :class="{ chosen: selected.status === item }" @click="changeStatus(item)">{{ statusText[item] }}</button></div><div class="fields"><label>单位<select v-model="selected.unit"><option v-for="unit in units" :key="unit.id" :value="unit.name">{{ unit.name }}</option></select></label><label>任务类型<select v-model="selected.taskType"><option value="code">代码</option><option value="document">文档</option><option value="support">支持</option></select></label><label>层级<select :value="selected.kind" @change="setKind(selected!, ($event.target as HTMLSelectElement).value as Kind)"><option v-for="(_, key) in kindText" :key="key" :value="key">{{ kindText[key as Kind] }}</option></select></label><label v-if="selected.kind === 'child'">所属主线<select v-model="selected.parentId"><option value="">请选择</option><option v-for="task in mains" :key="task.id" :value="task.id">{{ task.name }}</option></select></label><label>下发人<input v-model="selected.issuer" :disabled="selected.issuer === '本人'"><span class="inline-toggle"><input type="checkbox" :checked="selected.issuer === '本人'" @change="setSelf(selected!, $event)"><b>本人</b></span></label><label>发布日期<input v-model="selected.publishDate" type="date"></label><label>截止日期<input v-model="selected.dueDate" type="date" :disabled="!selected.dueDate"><span class="inline-toggle"><input type="checkbox" :checked="!selected.dueDate" @change="setUnlimited(selected!, $event)"><b>无期限</b></span></label><label v-if="selected.status === 'completed'">完成日期<input v-model="selected.completedAt" type="date"></label></div><section class="output-editor"><h3>产出</h3><label>结果<textarea v-model="selected.output.result"></textarea></label><div v-if="selected.taskType !== 'support'" class="quantity-grid"><template v-for="key in typeKeys[selected.taskType]" :key="key"><label>{{ outputLabels[key] }}<OutputStepper :model-value="selected.output[key]" @update:model-value="selected.output[key] = $event"/></label></template></div><p v-else class="support-hint">支持任务完成后自动计为 1 项支持。</p></section><button class="primary save" @click="save">保存全部修改</button></aside>
    <aside v-else class="details empty">选择一项任务</aside>

    <div v-if="showCreate" class="modal-backdrop"><form class="modal" @submit.prevent="create"><div class="modal-title"><h2>创建任务</h2><button class="close" type="button" @click="showCreate = false">×</button></div><div class="form-grid"><label>单位<select v-model="draft.unit" required><option v-for="unit in units" :key="unit.id" :value="unit.name">{{ unit.name }}</option></select></label><label>任务层级<select :value="draft.kind" @change="setKind(draft, ($event.target as HTMLSelectElement).value as Kind)"><option v-for="(_, key) in kindText" :key="key" :value="key">{{ kindText[key as Kind] }}</option></select></label><label>任务名<input v-model="draft.name" required></label><label>下发人<input v-model="draft.issuer" required :disabled="draft.issuer === '本人'"><span class="inline-toggle"><input type="checkbox" :checked="draft.issuer === '本人'" @change="setSelf(draft, $event)"><b>本人</b></span></label><label v-if="draft.kind === 'child'">所属主线<select v-model="draft.parentId" required><option value="">请选择</option><option v-for="task in mains" :key="task.id" :value="task.id">{{ task.name }}</option></select></label><label>发布日期<input v-model="draft.publishDate" type="date" required></label><label>截止日期<input v-model="draft.dueDate" type="date" :disabled="!draft.dueDate"><span class="inline-toggle"><input type="checkbox" :checked="!draft.dueDate" @change="setUnlimited(draft, $event)"><b>无期限</b></span></label><label>任务类型<select v-model="draft.taskType"><option value="code">代码</option><option value="document">文档</option><option value="support">支持</option></select></label></div><div class="actions"><button type="button" @click="showCreate = false">取消</button><button class="primary">创建任务</button></div></form></div>
    <div v-if="showUnits" class="modal-backdrop"><div class="modal"><div class="modal-title"><h2>单位列表</h2><button class="close" @click="showUnits = false">×</button></div><form class="unit-add" @submit.prevent="addUnit"><input v-model="newUnit" placeholder="输入单位名称"><button class="primary">新增单位</button></form><div class="unit-list"><div v-for="unit in units" :key="unit.id" class="unit-row"><span>{{ unit.name }}</span><em v-if="unit.id === defaultUnitID">默认单位</em><button @click="setDefault(unit.id)">设为默认</button><button @click="rename(unit)">修改</button></div></div></div></div>
    <button v-if="message" class="toast" @click="message = ''">{{ message }}</button>
    <div v-if="showAbout" class="modal-backdrop"><div class="modal"><div class="modal-title"><h2>关于</h2><button class="close" @click="showAbout = false">×</button></div><div class="about-grid"><div><small>程序名</small><strong>Task Manager | 工作任务管理器</strong></div><div><small>版本号</small><strong>alpha</strong></div><div><small>作者</small><strong>李正浩</strong></div></div></div></div>
  </main>
</template>
