type Runtime={importObject:WebAssembly.Imports;run:(instance:WebAssembly.Instance)=>Promise<void>};
const scope=globalThis as unknown as {
 Go:new()=>Runtime;
 pndDemoRequest:(path:string,body:string)=>string;
 postMessage:(value:unknown)=>void;
 onmessage:(event:MessageEvent)=>void;
};
let ready:Promise<void>;
async function initialize(base:string){
 const response=await fetch(new URL('manifest.json',base),{cache:'no-cache'});
 if(!response.ok)throw Error('Demo runtime manifest is unavailable');
 const manifest=await response.json();
 await import(/* @vite-ignore */ new URL(manifest.runtime,base).href);
 const runtime=new scope.Go();
 const module=await fetch(new URL(manifest.module,base));
 if(!module.ok||!module.body)throw Error('Demo codec is unavailable');
 const bytes=await new Response(module.body.pipeThrough(new DecompressionStream('gzip'))).arrayBuffer();
 const {instance}=await WebAssembly.instantiate(bytes,runtime.importObject);
 void runtime.run(instance).catch(error=>scope.postMessage({fatal:String(error)}));
 if(typeof scope.pndDemoRequest!=='function')throw Error('Demo codec did not initialize');
}
scope.onmessage=async({data})=>{
 try{
  if(data.initialize){ready=initialize(data.initialize);await ready;return}
  await ready;
  scope.postMessage({id:data.id,response:JSON.parse(scope.pndDemoRequest(data.path,data.body))});
 }catch(error){scope.postMessage({fatal:String(error)})}
};
export {};
