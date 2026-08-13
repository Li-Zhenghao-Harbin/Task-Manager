export namespace main {
	
	export class Output {
	    result: string;
	    newFeatures: number;
	    modifiedFeatures: number;
	    applications: number;
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

