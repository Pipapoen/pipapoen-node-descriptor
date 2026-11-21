import {chromium} from 'playwright';
import assert from 'node:assert/strict';
const base=process.env.PND_TEST_URL||'http://127.0.0.1:8080';
const browser=await chromium.launch({channel:process.env.PND_BROWSER_CHANNEL || (process.env.CI ? undefined : 'msedge'),headless:true});
try{
 const page=await browser.newPage({locale:'en-US'});const errors=[];page.on('pageerror',e=>errors.push(e.message));
 await page.goto(base);await page.locator('#node-id').waitFor();
 const metadata=await (await page.request.get(base+'/api/meta')).json();assert.equal(metadata.geoipEnabled,true);assert.equal(metadata.geoipAttribution,'DB-IP');
 for(const ip of ['8.8.8.8','1.1.1.1','2001:4860:4860::8888']){
  const response=await page.request.get(base+'/api/geoip?ip='+encodeURIComponent(ip));assert.equal(response.status(),200);const row=await response.json();assert.equal(row.source,'DB-IP');assert.equal(row.ip,ip);assert.equal(row.geo,0);assert.ok(row.country);assert.ok(row.name);assert.ok(Array.isArray(row.candidates));
 }
 assert.equal((await page.request.get(base+'/api/geoip?ip=127.0.0.1')).status(),400);
 await page.locator('#geo-tools > summary').click();await page.locator('#geo-tab-ip').click();
 await page.locator('#lookup-ip').fill('8.8.8.8');await page.locator('#geoip').click();await page.locator('#geoip-results .place-result').first().waitFor();
 assert.equal(await page.locator('#geo').inputValue(),'');assert.equal(await page.locator('#ipv4').inputValue(),'');
 const button=page.locator('#geoip-results .place-result').first();const id=(await button.textContent()).split(' · ').at(-1);await button.click();assert.equal(await page.locator('#geo').inputValue(),id);
 assert.equal(await page.locator('#ipv4').inputValue(),'');assert.equal(await page.getByRole('link',{name:'IP Geolocation by DB-IP'}).getAttribute('href'),'https://db-ip.com');
 await page.locator('#lookup-ip').fill('2001:4860:4860::8888');await page.locator('#geoip').click();await page.waitForFunction(()=>document.querySelector('#geoip-results').textContent.includes('2001:4860:4860::8888'));
 assert.deepEqual(errors,[]);console.log('Real DB-IP IPv4/IPv6, local candidates, explicit confirmation, and attribution passed');
}finally{await browser.close()}
