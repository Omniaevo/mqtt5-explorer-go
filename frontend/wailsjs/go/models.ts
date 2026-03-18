export namespace models {
	
	export class Connection {
	    id: number;
	    name: string;
	    mqttVersion: number;
	    protocol: string;
	    host: string;
	    port: number;
	    username?: string;
	    password?: string;
	    validateCert: boolean;
	    caFile?: string;
	    clientCert?: string;
	    clientKey?: string;
	    defaultSubscriptions?: string;
	    favourite: boolean;
	    // Go type: time
	    lastConnected: any;
	    // Go type: time
	    createdAt: any;
	    // Go type: time
	    updatedAt: any;
	
	    static createFrom(source: any = {}) {
	        return new Connection(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.name = source["name"];
	        this.mqttVersion = source["mqttVersion"];
	        this.protocol = source["protocol"];
	        this.host = source["host"];
	        this.port = source["port"];
	        this.username = source["username"];
	        this.password = source["password"];
	        this.validateCert = source["validateCert"];
	        this.caFile = source["caFile"];
	        this.clientCert = source["clientCert"];
	        this.clientKey = source["clientKey"];
	        this.defaultSubscriptions = source["defaultSubscriptions"];
	        this.favourite = source["favourite"];
	        this.lastConnected = this.convertValues(source["lastConnected"], null);
	        this.createdAt = this.convertValues(source["createdAt"], null);
	        this.updatedAt = this.convertValues(source["updatedAt"], null);
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
	export class Message {
	    id: number;
	    connectionId: number;
	    topic: string;
	    payload: number[];
	    qos: number;
	    retain: boolean;
	    // Go type: time
	    timestamp: any;
	    contentType?: string;
	    userProperties?: Record<string, string>;
	    responseTopic?: string;
	    correlationData?: number[];
	    messageExpiry?: number;
	    topicAlias?: number;
	    clientId?: string;
	
	    static createFrom(source: any = {}) {
	        return new Message(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.connectionId = source["connectionId"];
	        this.topic = source["topic"];
	        this.payload = source["payload"];
	        this.qos = source["qos"];
	        this.retain = source["retain"];
	        this.timestamp = this.convertValues(source["timestamp"], null);
	        this.contentType = source["contentType"];
	        this.userProperties = source["userProperties"];
	        this.responseTopic = source["responseTopic"];
	        this.correlationData = source["correlationData"];
	        this.messageExpiry = source["messageExpiry"];
	        this.topicAlias = source["topicAlias"];
	        this.clientId = source["clientId"];
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
	export class SendMessageRequest {
	    connectionId: number;
	    topic: string;
	    payload: string;
	    qos: number;
	    retain: boolean;
	    contentType?: string;
	    userProperties?: Record<string, string>;
	    responseTopic?: string;
	    correlationData?: string;
	    messageExpiry?: number;
	
	    static createFrom(source: any = {}) {
	        return new SendMessageRequest(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.connectionId = source["connectionId"];
	        this.topic = source["topic"];
	        this.payload = source["payload"];
	        this.qos = source["qos"];
	        this.retain = source["retain"];
	        this.contentType = source["contentType"];
	        this.userProperties = source["userProperties"];
	        this.responseTopic = source["responseTopic"];
	        this.correlationData = source["correlationData"];
	        this.messageExpiry = source["messageExpiry"];
	    }
	}
	export class TopicNode {
	    name: string;
	    fullTopic: string;
	    children?: Record<string, TopicNode>;
	    messageCount: number;
	    lastMessage?: Message;
	
	    static createFrom(source: any = {}) {
	        return new TopicNode(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.fullTopic = source["fullTopic"];
	        this.children = this.convertValues(source["children"], TopicNode, true);
	        this.messageCount = source["messageCount"];
	        this.lastMessage = this.convertValues(source["lastMessage"], Message);
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

}

