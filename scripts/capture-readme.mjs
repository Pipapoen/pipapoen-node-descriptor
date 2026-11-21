import { chromium } from 'playwright';
import { mkdir } from 'node:fs/promises';

const browser=await chromium.launch({channel:process.env.PND_BROWSER_CHANNEL || (process.env.CI ? undefined : 'msedge'),headless:true});
try {
 const page=await browser.newPage({viewport:{width:1280,height:1000},locale:'en-US',colorScheme:'light'});
 async function openPicker(kind){const other=kind==='country'?'city':'country';if(await page.locator('#'+other+'-picker').evaluate(el=>el.open))await page.locator('#'+other+'-filter').press('Escape');if(!await page.locator('#'+kind+'-picker').evaluate(el=>el.open))await page.locator('#'+kind+'-trigger').click()}
async function filterPicker(kind,value){await openPicker(kind);if(kind==='city'){const response=page.waitForResponse(r=>r.url().includes('/api/cities?')&&new URL(r.url()).searchParams.get('q')===value);await page.locator('#'+kind+'-filter').fill(value);await response}else await page.locator('#'+kind+'-filter').fill(value)}
async function choosePicker(kind,value){await openPicker(kind);await page.locator('#'+kind).selectOption(value);if(await page.locator('#'+kind+'-picker').evaluate(el=>el.open))await page.locator('#'+kind).press('Enter')}
 await page.goto(process.env.PND_TEST_URL||'http://127.0.0.1:8080');
 await page.locator('#node-id').waitFor();
 const descriptor={nodeId:'0123456789abcdef0123',geo:5128581,provider:1,architecture:0,nodeType:1,cpu:3,ram:8,storage:2,ipv4:'192.0.2.10'};
 await page.locator('#import-file').setInputFiles({name:'example.json',mimeType:'application/json',buffer:Buffer.from(JSON.stringify({descriptor}))});
 await page.waitForFunction(()=>document.querySelector('#node-id').value==='0123456789abcdef0123');
 await page.locator('#domain').fill('example.org');await page.locator('#generate').click();
 await page.waitForFunction(()=>document.querySelector('#hostname').textContent.endsWith('.example.org'));
 const hostname=await page.locator('#hostname').textContent();
 await mkdir('docs/images',{recursive:true});
 await page.screenshot({path:'docs/images/encode.png',fullPage:true});
 await page.locator('#geo-tools > summary').click();
 await page.waitForFunction(()=>!document.querySelector('#country-filter').disabled);
 await filterPicker('country','United States');await choosePicker('country','US');
 await filterPicker('city','New York');
 await page.waitForFunction(()=>Array.from(document.querySelector('#city').options).some(o=>o.value==='5128581'));
 await choosePicker('city','5128581');await page.locator('#use-city').click();
 await openPicker('city');
 await page.locator('.geo-field').scrollIntoViewIfNeeded();
 const field=await page.locator('.geo-field').boundingBox();
 const popup=await page.locator('#city-picker .picker-popup').boundingBox();
 const scrollY=await page.evaluate(()=>window.scrollY);
 await page.screenshot({path:'docs/images/location.png',clip:{x:Math.max(0,field.x-4),y:Math.max(0,field.y+scrollY-4),width:field.width+8,height:Math.max(field.y+field.height,popup.y+popup.height)-field.y+8}});
 await page.locator('[data-tab=decode]').click();await page.locator('#decode-input').fill(hostname);
 await page.locator('#decode-form button').click();await page.locator('#decoded').waitFor({state:'visible'});
 await page.waitForFunction(()=>document.querySelector('#decoded-city').textContent.includes('New York'));
 await page.screenshot({path:'docs/images/decode.png',fullPage:true});
 console.log('README screenshots captured with synthetic example data');
} finally {await browser.close()}
