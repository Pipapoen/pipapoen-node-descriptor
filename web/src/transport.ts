export type APIResult = {ok:boolean;result:any};

export async function request(path:string,data?:unknown):Promise<APIResult>{
 const response=await fetch(path,{method:data===undefined?'GET':'POST',headers:data===undefined?{}:{'Content-Type':'application/json'},body:data===undefined?undefined:JSON.stringify(data)});
 return {ok:response.ok,result:await response.json()};
}
