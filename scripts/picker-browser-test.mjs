import {chromium} from 'playwright';
import assert from 'node:assert/strict';
import {mkdir} from 'node:fs/promises';
import {pick} from './browser-select.mjs';
const browser=await chromium.launch({channel:process.env.PND_BROWSER_CHANNEL||(process.env.CI?undefined:'msedge'),headless:true});
try{
 const page=await browser.newPage({viewport:{width:1280,height:1000},locale:'en-US'});
 const errors=[];page.on('pageerror',error=>errors.push(error.message));
 await page.goto(process.env.PND_TEST_URL||'http://127.0.0.1:8080');await page.locator('#node-id').waitFor();
 assert.equal(await page.locator('select').count(),await page.locator('.choice-picker').count());
 for(const [id,values] of [['architecture',['3']],['nodeType',['7']],['provider',['0','65535']],['theme',['system']]]){
  assert.deepEqual(await page.locator('#'+id+' optgroup').first().locator('option').evaluateAll(options=>options.map(o=>o.value)),values);
 }
 await page.locator('#generate').click();await page.waitForFunction(()=>document.querySelector('#hostname').textContent);
 await page.locator('#architecture-trigger').click();await page.locator('#architecture-filter').fill('arm');
 assert.equal(await page.locator('#architecture').inputValue(),'');assert.equal(await page.locator('#stale').isVisible(),false);
 await page.locator('#architecture-filter').press('ArrowDown');await page.locator('#architecture').press('ArrowDown');await page.locator('#architecture').press('Enter');
 assert.equal(await page.locator('#architecture').inputValue(),'1');assert.equal(await page.locator('#architecture-picker').evaluate(p=>p.open),false);
 await page.locator('#architecture-trigger').press('Enter');await page.locator('#architecture-filter').press('Escape');
 assert.equal(await page.locator('#architecture-trigger').evaluate(e=>document.activeElement===e),true);
 await pick(page,'architecture','3');await pick(page,'nodeType','7');await pick(page,'ram','0');
 await page.locator('.recommendations > summary').click();
 for(const theme of ['light','dark']){
  await pick(page,'theme',theme);
  const styles=await page.evaluate(()=>['provider','architecture','nodeType','cpu','ram','storage','mode','plan-family','language','theme'].map(id=>{
   const s=getComputedStyle(document.getElementById(id+'-trigger'));
   return {background:s.backgroundColor,border:s.borderTopColor,radius:s.borderRadius,font:s.fontSize,height:s.height};
  }));
  for(const style of styles)assert.deepEqual(style,styles[0]);
 }
 await page.locator('#geo-tools > summary').click();
 assert.equal(await page.locator('#city-trigger').getAttribute('aria-disabled'),'true');
 await page.locator('#city-trigger').focus();await page.locator('#city-trigger').press('ArrowDown');
 assert.equal(await page.locator('#city-picker').evaluate(p=>p.open),false);
 await pick(page,'language','zh-TW');
 assert.equal(await page.locator('#architecture').inputValue(),'3');assert.equal(await page.locator('#nodeType').inputValue(),'7');assert.equal(await page.locator('#ram').inputValue(),'0');
 await page.locator('#architecture-trigger').click();
 await mkdir('build/test-results',{recursive:true});await page.screenshot({path:'build/test-results/unified-picker-desktop.png',fullPage:true,animations:'disabled'});
 await page.locator('#architecture-filter').press('Escape');
 await page.setViewportSize({width:390,height:844});await page.evaluate(()=>scrollTo(0,0));await page.locator('#theme-trigger').click();
 const popup=await page.locator('#theme-popup').boundingBox();assert.ok(popup.x>=0&&popup.x+popup.width<=390);
 assert.ok(await page.evaluate(()=>document.documentElement.scrollWidth<=innerWidth));
 await page.screenshot({path:'build/test-results/unified-picker-mobile.png',fullPage:true,animations:'disabled'});
 await page.locator('#theme-filter').press('Escape');
 assert.deepEqual(errors,[]);
 console.log('Unified selectors passed: shared surfaces, special groups, search, keyboard selection, Escape, disabled states, locale retention and mobile popup bounds.');
}finally{await browser.close()}
