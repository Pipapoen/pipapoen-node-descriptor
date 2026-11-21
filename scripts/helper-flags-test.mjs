import {chromium} from 'playwright';
import assert from 'node:assert/strict';

const base=process.env.PND_TEST_URL||'http://127.0.0.1:8080';
const browser=await chromium.launch({channel:process.env.PND_BROWSER_CHANNEL || (process.env.CI ? undefined : 'msedge'),headless:true});
try {
 for(let mask=0;mask<8;mask++){
  const page=await browser.newPage();const errors=[];page.on('pageerror',e=>errors.push(e.message));
  await page.route('**/api/meta',async route=>{const response=await route.fetch();const data=await response.json();await route.fulfill({response,json:{...data,countriesEnabled:!!(mask&1),placesEnabled:!!(mask&2),geoipEnabled:!!(mask&4)}})});
  await page.goto(base);await page.locator('#node-id').waitFor();
  assert.equal(await page.locator('#geo').isVisible(),true);
  assert.equal(await page.locator('#geo-tools').isVisible(),mask!==0);
  if(mask){
   await page.locator('#geo-tools > summary').click();
   const names=['browse','place','ip'].filter((_,i)=>mask&(1<<i));
   await page.locator('#geo-tab-'+names[0]).waitFor({state:'visible'});
   assert.equal(await page.locator('[data-geo-tab]:visible').count(),names.length);
   assert.equal(await page.locator('#geo-panel-'+names[0]).isVisible(),true);
   await page.locator('#geo-tab-'+names[0]).focus();await page.keyboard.press('ArrowRight');
   assert.equal(await page.locator('#geo-panel-'+names[1%names.length]).isVisible(),true);
  }else{
   await page.locator('#geo').fill('5128581');await page.locator('#generate').click();await page.waitForFunction(()=>document.querySelector('#hostname').textContent);
  }
  assert.deepEqual(errors,[]);await page.close();
 }
 console.log('All eight helper-visibility combinations passed (metadata mocked; backend gates tested in Go)');
} finally{await browser.close()}
