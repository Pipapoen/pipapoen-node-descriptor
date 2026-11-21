// Start the built application for browser tests and always stop it afterwards.
import {spawn} from 'node:child_process';
import {mkdtemp, writeFile, rm} from 'node:fs/promises';
import {tmpdir} from 'node:os';
import {join} from 'node:path';
import {randomBytes} from 'node:crypto';

const dir=await mkdtemp(join(tmpdir(),'pnd-browser-'));
const key=join(dir,'key');
await writeFile(key,randomBytes(32).toString('hex'));
const app=spawn(process.platform==='win32'?'build/pnd.exe':'build/pnd',[],{
 env:{...process.env,PND_KEY_FILE:key,PND_LISTEN:'127.0.0.1:8080',PND_GEONAMES_USERNAME:'',PND_GEOIP_DB:'',PND_OIDC_ISSUER:'',PND_OIDC_CLIENT_ID:'',PND_OIDC_CLIENT_SECRET_FILE:''},stdio:'inherit'
});
const stopped=new Promise(resolve=>app.once('exit',resolve));
try{
 let ready=false;
 for(let i=0;i<100;i++){
  if(app.exitCode!==null)throw Error('Application exited before readiness');
  try{const response=await fetch('http://127.0.0.1:8080/healthz');if(response.ok){ready=true;break}}catch{}
  await new Promise(resolve=>setTimeout(resolve,100));
 }
 if(!ready)throw Error('Application did not become ready');
 for(const script of ['browser-test.mjs','helper-flags-test.mjs']){
  const child=spawn(process.execPath,['scripts/'+script],{stdio:'inherit',env:{...process.env,PND_TEST_URL:'http://127.0.0.1:8080'}});
  const code=await new Promise(resolve=>child.once('exit',resolve));
  if(code!==0)throw Error(script+' failed');
 }
}finally{
 if(app.exitCode===null)app.kill();
 await stopped;
 await rm(dir,{recursive:true,force:true});
}
