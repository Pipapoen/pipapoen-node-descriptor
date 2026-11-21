import {pick} from './browser-select.mjs';
import {chromium} from 'playwright';
import assert from 'node:assert/strict';
import {mkdir} from 'node:fs/promises';
const base=process.env.PND_TEST_URL||'http://127.0.0.1:8080';
const browser=await chromium.launch({channel:process.env.PND_BROWSER_CHANNEL||(process.env.CI?undefined:'msedge'),headless:true});
try{
 const page=await browser.newPage({viewport:{width:1360,height:1000},locale:'en-US'});
 const errors=[];page.on('pageerror',error=>errors.push(error.message));
 await page.goto(base);await page.locator('#node-id').waitFor();
 const meta=await (await page.request.get(base+'/api/meta')).json();
 assert.ok(meta.providers.length>=40);assert.equal(meta.providers.length,meta.catalog.providers);assert.equal(meta.plans.length,meta.catalog.plans);
 assert.equal(meta.providers.find(p=>p.id===17)?.name,'OneProvider');assert.equal(meta.providers.find(p=>p.id===18)?.name,'DediRock');
 for(const theme of ['dark','light']){
  await pick(page,'theme',theme);
  await page.waitForFunction(()=>getComputedStyle(document.querySelector('#provider-trigger')).color===getComputedStyle(document.querySelector('#architecture-trigger')).color);
  const styles=await page.evaluate(()=>{
   const fields=['backgroundColor','borderTopColor','borderRadius','fontSize','height','color'];
   const read=selector=>Object.fromEntries(fields.map(key=>[key,getComputedStyle(document.querySelector(selector))[key]]));
   return {provider:read('#provider-trigger'),native:read('#architecture-trigger'),summary:getComputedStyle(document.querySelector('.recommendations>summary')).borderTopStyle};
  });
  assert.deepEqual(styles.provider,styles.native,`${theme} provider styling differs`);assert.equal(styles.summary,'solid');
 }
 const families=await page.locator('#plan-family optgroup').evaluateAll(groups=>groups.map(g=>({label:g.label,options:[...g.children].map(o=>({text:o.textContent,value:o.value}))})));
 assert.deepEqual(families.map(g=>g.label),['Product series','Reference configurations']);
 assert.ok(families[0].options.some(o=>o.text==='Scaleway \u00b7 BASIC3-X'));
 assert.ok(families[1].options.some(o=>o.text==='Oracle Cloud \u00b7 A1 Flex examples'));
 assert.ok(families[1].options.some(o=>o.text==='Self-managed \u00b7 Arm systems'));
 for(const group of families){
  const names=group.options.map(o=>o.text);
  assert.deepEqual(names,[...names].sort(new Intl.Collator('en',{sensitivity:'base',numeric:true}).compare));
 }
 const id=await page.locator('#node-id').inputValue();
 async function chooseProvider(value,query=''){
  if(!await page.locator('#provider-picker').evaluate(el=>el.open))await page.locator('#provider-trigger').click();
  await page.locator('#provider-filter').fill(query);
  await page.locator('#provider').selectOption(value);if(await page.locator('#provider-picker').evaluate(el=>el.open))await page.locator('#provider').press('Enter');
 }
 await page.locator('#geo').fill('5128581');await page.locator('#ipv4').fill('192.0.2.42');
 await chooseProvider('5','hetzner');
 assert.ok((await page.locator('#plan-family option').allTextContents()).includes('CAX (Arm)'));
 assert.ok(!(await page.locator('#plan-family option').allTextContents()).includes('BASIC3-X'));
 await page.locator('.recommendations > summary').click();
 await pick(page,'plan-family',JSON.stringify([5,'CAX (Arm)',true]));
 await page.locator('#plan-query').fill('CAX11');
 assert.equal(await page.locator('.suggestion').count(),1);
 await page.locator('.suggestion[data-plan="CAX11"]').click();
 assert.equal(await page.locator('#architecture').inputValue(),'1');
 assert.equal(await page.locator('#cpu').inputValue(),'1');
 assert.equal(await page.locator('#ram').inputValue(),'4');
 assert.equal(await page.locator('#storage').inputValue(),'1');
 assert.equal(await page.locator('#node-id').inputValue(),id);
 assert.equal(await page.locator('#geo').inputValue(),'5128581');
 assert.equal(await page.locator('#ipv4').inputValue(),'192.0.2.42');
 await chooseProvider('10','gcp');
 await pick(page,'storage','6');
 await page.locator('#plan-query').fill('e2-standard-2');
 assert.ok((await page.locator('.plan-row').textContent()).includes('Storage configured separately'));
 await page.locator('.suggestion[data-plan="e2-standard-2"]').click();
 assert.equal(await page.locator('#storage').inputValue(),'6');
 assert.equal(await page.locator('#architecture').inputValue(),'0');
 await page.locator('#generate').click();await page.waitForFunction(()=>document.querySelector('#hostname').textContent);
 await page.locator('#plan-query').fill('e2-standard-4');
 assert.equal(await page.locator('#stale').isVisible(),false);
 await pick(page,'ram','15');
 await chooseProvider('1','digital');
 assert.equal(await page.locator('#ram').inputValue(),'15');
 assert.equal(await page.locator('#ram option').count(),17);
 assert.equal(await page.locator('.suggestion').count(),12);
 await page.locator('#plan-more').click();assert.equal(await page.locator('.suggestion').count(),24);
 await page.locator('#plan-match').click();assert.equal(await page.locator('.suggestion').count(),0);
 await page.locator('#plan-match').click();
 const unknown={nodeId:id,provider:50000,cpu:0,ram:0};
 await page.locator('#import-file').setInputFiles({name:'unknown.json',mimeType:'application/json',buffer:Buffer.from(JSON.stringify({descriptor:unknown}))});
 await page.waitForFunction(()=>document.querySelector('#provider').value==='50000');
 assert.ok((await page.locator('#provider-label').textContent()).includes('50000'));
 assert.equal(await page.locator('#ram').inputValue(),'0');
 const response=page.waitForResponse(r=>r.url().endsWith('/api/encode'));
 await page.locator('#generate').click();const encoded=await (await response).json();assert.deepEqual(encoded.descriptor,unknown);
 const decoded=await page.request.post(base+'/api/decode',{data:{input:encoded.hostname}});assert.deepEqual(await decoded.json(),unknown);
 await chooseProvider('17','onecloud');
 await page.locator('#plan-query').fill('');assert.ok(await page.locator('.suggestion').count()>0);
 await chooseProvider('18','DEDIROCK');assert.ok(await page.locator('.suggestion').count()>0);
 await chooseProvider('19','racknerd');assert.equal(await page.locator('.suggestion').count(),0);
 assert.ok(await page.locator('.catalog-directory-link').isVisible());
 const directoryOutput=page.waitForResponse(r=>r.url().endsWith('/api/encode'));
 await page.locator('#generate').click();assert.equal((await (await directoryOutput).json()).descriptor.provider,19);
 await chooseProvider('18','dedirock');
 await pick(page,'theme','dark');
 await mkdir('build/test-results',{recursive:true});await page.screenshot({path:'build/test-results/catalog-desktop.png',fullPage:true,animations:'disabled'});
 await page.setViewportSize({width:390,height:844});
 const overflow=await page.evaluate(()=>[...document.querySelectorAll('body *')].map(e=>({tag:e.tagName,id:e.id,class:e.className,right:e.getBoundingClientRect().right,width:e.getBoundingClientRect().width})).filter(e=>e.width>0&&e.right>innerWidth+1));
 assert.equal(overflow.length,0,JSON.stringify(overflow.slice(0,12)));
 await page.screenshot({path:'build/test-results/catalog-mobile.png',fullPage:true,animations:'disabled'});
 await pick(page,'language','zh-TW');assert.equal(await page.locator('#provider').inputValue(),'18');
 assert.deepEqual(errors,[]);
 console.log('Catalog browser checks passed: aliases, families, named plans, Arm, optional disks, filters, pagination, retained fields, unknown IDs and mobile layout.');
}finally{await browser.close()}
