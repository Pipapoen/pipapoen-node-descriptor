import {chromium} from 'playwright';
import {createServer} from 'node:http';
import {readFile,mkdir} from 'node:fs/promises';
import {resolve,extname} from 'node:path';
import assert from 'node:assert/strict';
const root=resolve('dist/demo');
const server=createServer(async(req,res)=>{
 const path=decodeURIComponent(new URL(req.url,'http://localhost').pathname).replace(/^\/pnd-demo\//,'');
 const file=resolve(root,path||'index.html');
 if(!file.startsWith(root+'/')&&!file.startsWith(root+'\\')){res.writeHead(404);res.end();return}
 try{res.setHeader('Content-Type',({'.html':'text/html; charset=utf-8','.js':'text/javascript','.css':'text/css','.json':'application/json'})[extname(file)]||'application/octet-stream');res.end(await readFile(file))}catch{res.writeHead(404);res.end()}
});
await new Promise(resolve=>server.listen(0,'127.0.0.1',resolve));
const browser=await chromium.launch({channel:process.env.PND_BROWSER_CHANNEL||(process.env.CI?undefined:'msedge'),headless:true});
try{
 const page=await browser.newPage({viewport:{width:1280,height:1000},locale:'en-US'});
 const errors=[];const apiRequests=[];
 page.on('pageerror',error=>errors.push(error.message));
 page.on('request',request=>{if(request.url().includes('/api/'))apiRequests.push(request.url())});
 const base=`http://127.0.0.1:${server.address().port}/pnd-demo/`;
 await page.context().grantPermissions(['clipboard-read','clipboard-write'],{origin:new URL(base).origin});
 await page.goto(base);
 await page.locator('#node-id').waitFor();
 const original=await page.locator('#node-id').inputValue();
 assert.match(original,/^[0-9a-f]{20}$/);
 assert.equal(await page.locator('#node-id').getAttribute('readonly'),'');
 assert.equal(await page.locator('#encode-form').count(),1);
 assert.equal(await page.locator('#decode-form').count(),1);
 assert.ok((await page.locator('.version-label').textContent()).includes('public key'));

 // Native Go fixtures verify real WASM ciphertext parity through the production form.
 const fixtures=JSON.parse(await readFile('scripts/testdata/demo.json','utf8'));
 async function importNode(descriptor){
  await page.locator('#import-file').setInputFiles({name:'node.json',mimeType:'application/json',buffer:Buffer.from(JSON.stringify({descriptor}))});
  await page.waitForFunction(d=>document.querySelector('#node-id').value===d.nodeId&&document.querySelector('#provider').value===String(d.provider??'')&&document.querySelector('#geo').value===String(d.geo??''),descriptor);
 }
 async function generate(expected){
  await page.locator('#generate').click();
  if(expected)await page.waitForFunction(value=>document.querySelector('#hostname').textContent===value,expected);
  else await page.waitForFunction(()=>document.querySelector('#hostname').textContent&&document.querySelector('#stale').hidden);
 }
 for(const row of fixtures){
  await importNode(row.descriptor);
  for(const mode of ['compact','padded']){
   await page.locator('#mode').selectOption(mode);await generate(row[mode].hostname);
   assert.equal(await page.locator('#hostname').textContent(),row[mode].hostname);
  }
 }

 // Free-form inputs, custom prefix/domain and IPv6 operate beyond the old presets.
 const custom={nodeId:original,geo:5128581,provider:1,architecture:0,nodeType:1,cpu:3,ram:8,storage:2,ipv4:'192.0.2.42',ipv6:'2001:db8::42'};
 await importNode(custom);
 await page.locator('#prefix').fill('z');await page.locator('#domain').fill('nodes.example.org');
 await page.locator('#mode').selectOption('compact');await generate();
 const compact=await page.locator('#hostname').textContent();
 assert.match(compact,/^z0[a-z0-9]+\.nodes\.example\.org$/);
 await page.locator('#mode').selectOption('padded');await generate();
 const hostname=await page.locator('#hostname').textContent();
 assert.equal(hostname.split('.')[0].length,63);
 await page.locator('#copy').click();await page.waitForFunction(()=>document.querySelector('#copy').textContent==='Copied');
 assert.equal(await page.evaluate(()=>navigator.clipboard.readText()),hostname);
 await generate();await page.waitForFunction(()=>document.querySelector('#copy').textContent==='Copy');
 await page.locator('[data-tab=decode]').click();await page.locator('#decode-input').fill('https://'+hostname+'/');
 await page.locator('#decode-form button').click();await page.locator('#decoded').waitFor({state:'visible'});
 await page.waitForFunction(()=>document.querySelector('#decoded-city').textContent.includes('New York'));
 assert.ok((await page.locator('#decoded-values').textContent()).includes('2001:db8::42'));
 assert.ok((await page.locator('#decoded-values').textContent()).includes(original));
 const download=page.waitForEvent('download');await page.locator('#export-decoded').click();
 const saved=await download;const contents=JSON.parse(await readFile(await saved.path(),'utf8'));
 assert.deepEqual(contents.descriptor,custom);
 await page.locator('#edit-decoded').click();assert.equal(await page.locator('#node-id').inputValue(),original);

 await page.locator('#geo-tools > summary').click();
 await page.waitForFunction(()=>!document.querySelector('#country-filter').disabled);
 await page.locator('#country-trigger').click();await page.locator('#country-filter').fill('United States');
 await page.locator('#country').selectOption('US');await page.locator('#country').press('Enter');
 await page.locator('#city-trigger').click();await page.locator('#city-filter').fill('New York');
 await page.waitForFunction(()=>Array.from(document.querySelector('#city').options).some(o=>o.value==='5128581'));
 await page.locator('#city').selectOption('5128581');await page.locator('#city').press('Enter');await page.locator('#use-city').click();
 assert.equal(await page.locator('#geo').inputValue(),'5128581');
 assert.equal(await page.locator('#geo-tab-place').isVisible(),false);
 assert.equal(await page.locator('#geo-tab-ip').isVisible(),false);
 await page.locator('#geo-tools > summary').click();
 await page.locator('#language').selectOption('zh-TW');assert.equal(await page.locator('html').getAttribute('lang'),'zh-TW');
 assert.equal(await page.locator('#node-id').inputValue(),original);
 assert.equal(await page.locator('#domain').inputValue(),'nodes.example.org');
 await page.locator('#theme').selectOption('dark');assert.equal(await page.locator('html').getAttribute('data-theme'),'dark');
 await page.locator('#theme').selectOption('system');await page.emulateMedia({colorScheme:'light'});
 await page.waitForFunction(()=>document.documentElement.dataset.theme==='light');
 await page.locator('#language').selectOption('en');
 await page.locator('#mode').selectOption('compact');await generate();
 await mkdir('test-results',{recursive:true});await page.screenshot({path:'test-results/demo-desktop.png',fullPage:true});
 await page.setViewportSize({width:390,height:844});
 assert.equal(await page.evaluate(()=>document.documentElement.scrollWidth<=innerWidth),true);
 await page.screenshot({path:'test-results/demo-mobile.png',fullPage:true});
 page.once('dialog',dialog=>dialog.accept());await page.locator('#new-id').click();
 await page.waitForFunction(id=>document.querySelector('#node-id').value!==id,original);
 await page.locator('[data-tab=decode]').click();await page.locator('#decode-input').fill('invalid');await page.locator('#decode-form button').click();
 await page.locator('#notice').waitFor({state:'visible'});
 assert.deepEqual(errors,[]);assert.deepEqual(apiRequests,[]);
 console.log('Shared demo passed: native ciphertext parity, free-form encoding/decoding, import/export, geography, clipboard, locales/themes, mobile layout, and zero API requests.');
}finally{await browser.close();await new Promise(resolve=>server.close(resolve))}
