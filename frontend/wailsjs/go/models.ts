export namespace main {
	
	export class AgentApp {
	    id: string;
	    name: string;
	    model: string;
	    taskIds: string[];
	    available: boolean;
	    resetAt: string;
	    prompts: Record<string, Array<AgentPrompt>>;
	    createdAt: string;
	    updatedAt: string;
	
	    static createFrom(source: any = {}) {
	        return new AgentApp(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.name = source["name"];
	        this.model = source["model"];
	        this.taskIds = source["taskIds"];
	        this.available = source["available"];
	        this.resetAt = source["resetAt"];
	        this.prompts = this.convertValues(source["prompts"], Array<AgentPrompt>, true);
	        this.createdAt = source["createdAt"];
	        this.updatedAt = source["updatedAt"];
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class AgentAppInput {
	    name: string;
	    model: string;
	    taskIds: string[];
	    available: boolean;
	    resetAt: string;
	
	    static createFrom(source: any = {}) {
	        return new AgentAppInput(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.model = source["model"];
	        this.taskIds = source["taskIds"];
	        this.available = source["available"];
	        this.resetAt = source["resetAt"];
	    }
	}
	export class AgentPrompt {
	    id: string;
	    text: string;
	
	    static createFrom(source: any = {}) {
	        return new AgentPrompt(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.text = source["text"];
	    }
	}
	export class Output {
	    result: string;
	    newFeatures: number;
	    modifiedFeatures: number;
	    applications: number;
	    skills: number;
	    newDocuments: number;
	    modifiedDocuments: number;
	    supports: number;
	
	    static createFrom(source: any = {}) {
	        return new Output(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.result = source["result"];
	        this.newFeatures = source["newFeatures"];
	        this.modifiedFeatures = source["modifiedFeatures"];
	        this.applications = source["applications"];
	        this.skills = source["skills"];
	        this.newDocuments = source["newDocuments"];
	        this.modifiedDocuments = source["modifiedDocuments"];
	        this.supports = source["supports"];
	    }
	}
	export class Settings {
	    backgroundColor: string;
	    autoStart: boolean;
	
	    static createFrom(source: any = {}) {
	        return new Settings(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.backgroundColor = source["backgroundColor"];
	        this.autoStart = source["autoStart"];
	    }
	}
	export class Task {
	    id: string;
	    name: string;
	    kind: string;
	    parentId: string;
	    issuer: string;
	    unit: string;
	    publishDate: string;
	    dueDate: string;
	    taskType: string;
	    status: string;
	    deliverable: string;
	    agentId: string;
	    createdAt: string;
	    updatedAt: string;
	    completedAt: string;
	    output: Output;
	
	    static createFrom(source: any = {}) {
	        return new Task(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.name = source["name"];
	        this.kind = source["kind"];
	        this.parentId = source["parentId"];
	        this.issuer = source["issuer"];
	        this.unit = source["unit"];
	        this.publishDate = source["publishDate"];
	        this.dueDate = source["dueDate"];
	        this.taskType = source["taskType"];
	        this.status = source["status"];
	        this.deliverable = source["deliverable"];
	        this.agentId = source["agentId"];
	        this.createdAt = source["createdAt"];
	        this.updatedAt = source["updatedAt"];
	        this.completedAt = source["completedAt"];
	        this.output = this.convertValues(source["output"], Output);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class TaskInput {
	    name: string;
	    kind: string;
	    parentId: string;
	    issuer: string;
	    unit: string;
	    publishDate: string;
	    dueDate: string;
	    taskType: string;
	    agentId: string;
	
	    static createFrom(source: any = {}) {
	        return new TaskInput(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.kind = source["kind"];
	        this.parentId = source["parentId"];
	        this.issuer = source["issuer"];
	        this.unit = source["unit"];
	        this.publishDate = source["publishDate"];
	        this.dueDate = source["dueDate"];
	        this.taskType = source["taskType"];
	        this.agentId = source["agentId"];
	    }
	}
	export class Unit {
	    id: string;
	    name: string;
	
	    static createFrom(source: any = {}) {
	        return new Unit(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.name = source["name"];
	    }
	}

}

