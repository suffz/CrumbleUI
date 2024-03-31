export namespace apiGO {
	
	
	export class UserINFO {
	    id: string;
	    name: string;
	
	    static createFrom(source: any = {}) {
	        return new UserINFO(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.name = source["name"];
	    }
	}

}

export namespace main {
	
	export class Logs {
	    name: string;
	    content: apiGO.Details[];
	
	    static createFrom(source: any = {}) {
	        return new Logs(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.content = this.convertValues(source["content"], apiGO.Details);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice) {
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

}

export namespace utils {
	
	
	export class Tasks {
	    name: string;
	    start: number;
	    end: number;
	    headurl: string;
	    active: boolean;
	
	    static createFrom(source: any = {}) {
	        return new Tasks(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.start = source["start"];
	        this.end = source["end"];
	        this.headurl = source["headurl"];
	        this.active = source["active"];
	    }
	}

}

