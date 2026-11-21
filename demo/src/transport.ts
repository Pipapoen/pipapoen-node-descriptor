import type {APIResult} from '../../web/src/transport';

const worker=new Worker(new URL('./worker.ts',import.meta.url),{type:'module'});
let sequence=0;
const pending=new Map<number,{resolve:(value:APIResult)=>void;reject:(error:Error)=>void}>();
let failed:Error|undefined;
function fail(error:Error){failed=error;for(const request of pending.values())request.reject(error);pending.clear()}
worker.onerror=event=>fail(new Error(event.message||'Demo worker failed to load'));
worker.onmessage=({data})=>{
 if(data.fatal){fail(new Error(data.fatal));return}
 const request=pending.get(data.id);
 if(request){pending.delete(data.id);request.resolve(data.response)}
};
worker.postMessage({initialize:new URL('runtime/',new URL(import.meta.env.BASE_URL,document.baseURI)).href});

export function request(path:string,data?:unknown):Promise<APIResult>{
 if(failed)return Promise.reject(failed);
 return new Promise((resolve,reject)=>{const id=++sequence;pending.set(id,{resolve,reject});worker.postMessage({id,path,body:JSON.stringify(data??{})})});
}
