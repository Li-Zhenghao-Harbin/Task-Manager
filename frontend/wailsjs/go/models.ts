export namespace main {
	
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
	    summary: string;
	    createdAt: string;
	    updatedAt: string;
	
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
	        this.summary = source["summary"];
	        this.createdAt = source["createdAt"];
	        this.updatedAt = source["updatedAt"];
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

