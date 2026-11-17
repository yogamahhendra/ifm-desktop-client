export namespace repositories {
	
	export class APIKeyItem {
	    id: string;
	    name: string;
	    key_mask: string;
	    is_active: boolean;
	    total_prompt_tokens: number;
	    total_completion_tokens: number;
	    total_requests: number;
	    // Go type: time
	    last_used_at?: any;
	    // Go type: time
	    created_at: any;
	
	    static createFrom(source: any = {}) {
	        return new APIKeyItem(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.name = source["name"];
	        this.key_mask = source["key_mask"];
	        this.is_active = source["is_active"];
	        this.total_prompt_tokens = source["total_prompt_tokens"];
	        this.total_completion_tokens = source["total_completion_tokens"];
	        this.total_requests = source["total_requests"];
	        this.last_used_at = this.convertValues(source["last_used_at"], null);
	        this.created_at = this.convertValues(source["created_at"], null);
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
	export class Conversation {
	    id: string;
	    title: string;
	    model_id: string;
	    preset_id?: string;
	    session_id: string;
	    system_prompt: string;
	    folder_path?: string;
	    // Go type: time
	    created_at: any;
	    // Go type: time
	    updated_at: any;
	
	    static createFrom(source: any = {}) {
	        return new Conversation(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.title = source["title"];
	        this.model_id = source["model_id"];
	        this.preset_id = source["preset_id"];
	        this.session_id = source["session_id"];
	        this.system_prompt = source["system_prompt"];
	        this.folder_path = source["folder_path"];
	        this.created_at = this.convertValues(source["created_at"], null);
	        this.updated_at = this.convertValues(source["updated_at"], null);
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
	    id: string;
	    conversation_id: string;
	    parent_message_id?: string;
	    role: string;
	    content: string;
	    reasoning_content?: string;
	    tool_calls?: string;
	    tool_call_id?: string;
	    finish_reason?: string;
	    prompt_tokens: number;
	    completion_tokens: number;
	    // Go type: time
	    created_at: any;
	
	    static createFrom(source: any = {}) {
	        return new Message(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.conversation_id = source["conversation_id"];
	        this.parent_message_id = source["parent_message_id"];
	        this.role = source["role"];
	        this.content = source["content"];
	        this.reasoning_content = source["reasoning_content"];
	        this.tool_calls = source["tool_calls"];
	        this.tool_call_id = source["tool_call_id"];
	        this.finish_reason = source["finish_reason"];
	        this.prompt_tokens = source["prompt_tokens"];
	        this.completion_tokens = source["completion_tokens"];
	        this.created_at = this.convertValues(source["created_at"], null);
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
	export class ModelItem {
	    id: string;
	    name: string;
	    model_identifier: string;
	    description: string;
	    temperature: number;
	    top_p: number;
	    top_k?: number;
	    max_tokens: number;
	    reasoning_effort: string;
	    is_default: boolean;
	    // Go type: time
	    created_at: any;
	    // Go type: time
	    updated_at: any;
	
	    static createFrom(source: any = {}) {
	        return new ModelItem(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.name = source["name"];
	        this.model_identifier = source["model_identifier"];
	        this.description = source["description"];
	        this.temperature = source["temperature"];
	        this.top_p = source["top_p"];
	        this.top_k = source["top_k"];
	        this.max_tokens = source["max_tokens"];
	        this.reasoning_effort = source["reasoning_effort"];
	        this.is_default = source["is_default"];
	        this.created_at = this.convertValues(source["created_at"], null);
	        this.updated_at = this.convertValues(source["updated_at"], null);
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
	export class PresetItem {
	    id: string;
	    name: string;
	    description: string;
	    system_prompt: string;
	    model_id?: string;
	    temperature?: number;
	    top_p?: number;
	    max_tokens?: number;
	    reasoning_effort?: string;
	    // Go type: time
	    created_at: any;
	    // Go type: time
	    updated_at: any;
	
	    static createFrom(source: any = {}) {
	        return new PresetItem(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.name = source["name"];
	        this.description = source["description"];
	        this.system_prompt = source["system_prompt"];
	        this.model_id = source["model_id"];
	        this.temperature = source["temperature"];
	        this.top_p = source["top_p"];
	        this.max_tokens = source["max_tokens"];
	        this.reasoning_effort = source["reasoning_effort"];
	        this.created_at = this.convertValues(source["created_at"], null);
	        this.updated_at = this.convertValues(source["updated_at"], null);
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

export namespace settings {
	
	export class AppConfigResponse {
	    theme: string;
	    ifm_base_url: string;
	    default_model: string;
	    default_preset: string;
	    request_timeout: string;
	    streaming_enabled: string;
	    temperature: string;
	    top_p: string;
	    max_tokens: string;
	    reasoning_effort: string;
	    has_api_key: boolean;
	    has_tavily_key: boolean;
	    database_path: string;
	    all_settings: Record<string, string>;
	
	    static createFrom(source: any = {}) {
	        return new AppConfigResponse(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.theme = source["theme"];
	        this.ifm_base_url = source["ifm_base_url"];
	        this.default_model = source["default_model"];
	        this.default_preset = source["default_preset"];
	        this.request_timeout = source["request_timeout"];
	        this.streaming_enabled = source["streaming_enabled"];
	        this.temperature = source["temperature"];
	        this.top_p = source["top_p"];
	        this.max_tokens = source["max_tokens"];
	        this.reasoning_effort = source["reasoning_effort"];
	        this.has_api_key = source["has_api_key"];
	        this.has_tavily_key = source["has_tavily_key"];
	        this.database_path = source["database_path"];
	        this.all_settings = source["all_settings"];
	    }
	}

}

export namespace workspace {
	
	export class WorkspaceFile {
	    path: string;
	    relative_path: string;
	    size: number;
	    extension: string;
	    tokens: number;
	    included: boolean;
	
	    static createFrom(source: any = {}) {
	        return new WorkspaceFile(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.path = source["path"];
	        this.relative_path = source["relative_path"];
	        this.size = source["size"];
	        this.extension = source["extension"];
	        this.tokens = source["tokens"];
	        this.included = source["included"];
	    }
	}
	export class WorkspaceSummary {
	    folder_path: string;
	    folder_name: string;
	    total_files: number;
	    total_size: number;
	    total_tokens: number;
	    files: WorkspaceFile[];
	
	    static createFrom(source: any = {}) {
	        return new WorkspaceSummary(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.folder_path = source["folder_path"];
	        this.folder_name = source["folder_name"];
	        this.total_files = source["total_files"];
	        this.total_size = source["total_size"];
	        this.total_tokens = source["total_tokens"];
	        this.files = this.convertValues(source["files"], WorkspaceFile);
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

