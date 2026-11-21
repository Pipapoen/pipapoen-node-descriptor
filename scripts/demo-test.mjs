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
 const page=await browser.newPage({viewport:{width:1280,height:900}});
 const errors=[];const apiRequests=[];
 page.on('pageerror',error=>errors.push(error.message));
 page.on('request',request=>{if(request.url().includes('/api/'))apiRequests.push(request.url())});
 await page.goto(`http://127.0.0.1:${server.address().port}/pnd-demo/`);
 for(const node of ['0','1','2']){
  await page.locator('#node').selectOption(node);
  for(const mode of ['compact','padded']){
   await page.locator('#mode').selectOption(mode);await page.locator('#generate').click();
   const hostname=await page.locator('#hostname').textContent();
   assert.match(hostname,/^x0[a-z0-9]+\.example\.org$/);
   if(mode==='padded')assert.equal(hostname.split('.')[0].length,63);
   await page.locator('#decode').click();assert.ok(await page.locator('#decoded').isVisible());
  }
 }
 await page.locator('#language').selectOption('zh-TW');assert.equal(await page.locator('html').getAttribute('lang'),'zh-TW');
 await page.locator('#theme').selectOption('dark');assert.equal(await page.locator('html').getAttribute('data-theme'),'dark');
 await page.locator('#theme').selectOption('system');await page.emulateMedia({colorScheme:'light'});assert.equal(await page.locator('html').getAttribute('data-theme'),'light');
 await page.locator('#language').selectOption('en');
 await page.locator('#node').selectOption('1');await page.locator('#mode').selectOption('compact');await page.locator('#generate').click();await page.locator('#decode').click();
 await mkdir('test-results',{recursive:true});await page.screenshot({path:'test-results/demo-desktop.png',fullPage:true});
 await page.setViewportSize({width:375,height:812});
 assert.equal(await page.evaluate(()=>document.documentElement.scrollWidth<=innerWidth),true);
 await page.screenshot({path:'test-results/demo-mobile.png',fullPage:true});
 assert.deepEqual(errors,[]);assert.deepEqual(apiRequests,[]);
 console.log('Static demo passed: preset round trips, modes, locales, themes, subpath hosting, mobile layout, no API requests.');
}finally{await browser.close();await new Promise(resolve=>server.close(resolve))}
