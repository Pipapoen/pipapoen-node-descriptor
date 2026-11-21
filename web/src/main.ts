import '../../node_modules/basecoat-css/dist/basecoat-vega.cdn.css';
import './style.css';
import en from './locales/en.json';
import zh from './locales/zh-TW.json';
import {request} from '@pnd/transport';
import {setupCatalog,type Plan,type Provider} from './catalog';
import {setupPickers} from './pickers';

type Locale = typeof en;
type Descriptor = {nodeId:string; geo?:number; provider?:number; architecture?:number; nodeType?:number; cpu?:number; ram?:number; storage?:number; ipv4?:string; ipv6?:string};
type Meta={demo?:boolean;brandName:string;prefix:string;domain:string;mode:string;cpu:string[];ram:string[];storage:string[];providers:Provider[];plans:Plan[];catalog:{checkedAt:string;plans:number};placesEnabled:boolean;geoipEnabled:boolean;countriesEnabled:boolean;geoipAttribution:string};
type Preview={bits:number;capacityBits:number;compactLength:number;labelLength:number;mode:string};
const locales:Record<string,Locale>={'en':en,'zh-TW':zh};
let lang=localStorage.getItem('pnd.language') || (navigator.language.toLowerCase().startsWith('zh')?'zh-TW':'en');
if(!locales[lang])lang='en';
let theme=localStorage.getItem('pnd.theme')||'system';
if(!['light','dark','system'].includes(theme))theme='system';
let meta:Meta, nodeId='', lastDecoded:Descriptor|undefined, lastOutput:Descriptor|undefined, lastHostname='', outputStale=false, tab='encode', sequence=0, timer:ReturnType<typeof setTimeout>;
let session={required:false,authenticated:false,brandName:''};
let citySequence=0, cityOffset=0, mountGeneration=0;
let pickerUI:ReturnType<typeof setupPickers>|undefined;
let decodedLocationSequence=0;
let outputRevision=0;
let catalogUI:ReturnType<typeof setupCatalog>|undefined;
let cityTimer:ReturnType<typeof setTimeout>;
let cityRows:{id:number;name:string;admin:string;feature:string;names?:Record<string,string>;adminNames?:Record<string,string>}[]=[];
const $=<T extends HTMLElement>(s:string)=>document.querySelector<T>(s)!;
const t=(key:string):string=>(locales[lang] as Record<string,string>)[key]||(en as Record<string,string>)[key]||key;
const escape=(s:string)=>s.replace(/[&<>"']/g,c=>({'&':'&amp;','<':'&lt;','>':'&gt;','"':'&quot;',"'":'&#39;'}[c]!));
const text=(s:string)=>escape(t(s));
function applyTheme(){const resolved=theme==='system'?(matchMedia('(prefers-color-scheme: dark)').matches?'dark':'light'):theme;document.documentElement.dataset.theme=resolved;document.documentElement.classList.toggle('dark',resolved==='dark')}
applyTheme();matchMedia('(prefers-color-scheme: dark)').addEventListener('change',applyTheme);
async function api<T>(path:string,data?:unknown):Promise<T>{const {ok,result}=await request(path,data);if(!ok){if(result.error==='login_required'&&session.required)location.replace('/auth/login');throw new Error(t(result.error||'error'))}return result}
function help(key:string){return `<details class="help"><summary aria-label="${text('help')}">?</summary><p>${text(key)}</p></details>`}
function field(key:string,input:string){return `<div class="field"><label class="label" for="${key}">${text(key)}</label>${input}</div>`}
function select(key:string,options:{value:number;label:string}[]){
 const special=key==='architecture'?3:key==='nodeType'?7:undefined;
 const render=(rows:typeof options)=>rows.map(o=>`<option value="${o.value}">${escape(o.label)}</option>`).join('');
 const groups=special===undefined?render(options):`<optgroup label="${text('specialOptions')}">${render(options.filter(o=>o.value===special))}</optgroup><optgroup label="${text('standardOptions')}">${render(options.filter(o=>o.value!==special))}</optgroup>`;
 return `<select class="select" id="${key}" aria-label="${text(key)}"><option value="">${text('none')}</option>${groups}</select>`;
}
function number(key:string,max:number){return `<input class="input" id="${key}" type="number" min="1" max="${max}" step="1" inputmode="numeric" aria-label="${text(key)}" placeholder="—">`}
function optionList(v:string[]){return v.map((label,value)=>({label,value}))}
function current():Descriptor{
 const d:Descriptor={nodeId};
 for(const key of ['geo','provider','architecture','nodeType','cpu','ram','storage','ipv4','ipv6'] as const){
  const value=$<HTMLInputElement>(`#${key}`).value.trim();
  if(value!==''){if(key==='ipv4'||key==='ipv6')d[key]=value;else d[key]=Number(value)}
 }
 return d;
}
function settings(){return {prefix:$<HTMLInputElement>('#prefix').value,domain:$<HTMLInputElement>('#domain').value,mode:$<HTMLSelectElement>('#mode').value}}
function showError(error:unknown){$('#notice').textContent=error instanceof Error?error.message:t('error');$('#notice').hidden=false}
function clearError(){$('#notice').hidden=true}
function restore(d:Descriptor){
 nodeId=d.nodeId;$<HTMLInputElement>('#node-id').value=nodeId;
 for(const key of ['geo','provider','architecture','nodeType','cpu','ram','storage','ipv4','ipv6'] as const){
  const input=$<HTMLInputElement>(`#${key}`),value=d[key];
  if(value!==undefined&&input instanceof HTMLSelectElement&&!Array.from(input.options).some(o=>o.value===String(value)))input.add(new Option(String(value),String(value)));
  input.value=value===undefined?'':String(value);
 }
 recommend();pickerUI?.refresh();schedulePreview();
}
function mount(){
 ++mountGeneration;++citySequence;clearTimeout(cityTimer);pickerUI?.destroy();
 document.documentElement.lang=lang;
 $('#app').innerHTML=`
 <div class="app-content">
  <header class="app-header">
   <button class="brand-link" id="brand-home" type="button"><span class="brand-symbol" aria-hidden="true"></span><strong>PND</strong>${meta.brandName?`<span class="brand-name">${escape(meta.brandName)}</span>`:''}</button>
   <nav class="workspace-nav" aria-label="Workspace">${['encode','decode'].map(k=>`<button class="nav-link ${tab===k?'active':''}" data-tab="${k}">${text(k)}</button>`).join('')}</nav>
   <div class="toolbar"><select class="select" id="language" aria-label="${text('language')}"><option value="en">English</option><option value="zh-TW">${escape(zh.localeName)}</option></select><select class="select" id="theme" aria-label="${text('theme')}"><optgroup label="${text('specialOptions')}"><option value="system">${text('system')}</option></optgroup><optgroup label="${text('standardOptions')}">${['light','dark'].map(k=>`<option value="${k}">${text(k)}</option>`).join('')}</optgroup></select>${session.required?`<button id="logout" data-variant="outline" class="btn btn-outline btn-sm">${text('logout')}</button>`:''}</div>
  </header>
  <main>
   <div class="page-heading"><h1>${text(tab==='encode'?'builder':'decode')}</h1><span class="text-secondary version-label">${text(meta.demo?'demoLabel':'experiment')}${meta.demo?help('demoHelp'):''}</span></div>
   <p id="notice" class="alert alert-warning" role="alert" hidden></p>
   <div id="encode-panel" class="workspace">
    <form id="encode-form" class="card editor-card">
     <div class="card-header"><h2 class="card-title">${text('nodeDetails')}</h2><div class="card-actions"><button type="button" id="import" data-variant="outline" class="btn btn-outline btn-sm">${text('import')}</button><button type="button" id="export-input" data-variant="outline" class="btn btn-outline btn-sm">${text('export')}</button><input type="file" id="import-file" accept="application/json,.json" hidden></div></div>
     <div class="card-body">
      <section class="identity-block"><div class="section-heading"><label class="label" for="node-id">Node ID</label>${help('identityHelp')}</div><div class="input-group"><input id="node-id" class="input mono" readonly spellcheck="false" aria-readonly="true"><button type="button" id="new-id" data-variant="outline" class="btn btn-outline">${text('regenerate')}</button></div></section>
      <section><h2 class="section-label">${text('placement')}</h2><div class="placement-grid"><div class="field"><label class="label" for="provider">${text('provider')}</label><details class="search-picker" id="provider-picker"><summary id="provider-trigger"><span id="provider-label">${text('none')}</span></summary><div class="picker-popup"><label class="sr-only" for="provider-filter">${text('providerFilter')}</label><input class="input" id="provider-filter" type="search" placeholder="${text('providerFilter')}" autocomplete="off"><select class="picker-list" id="provider" size="7" aria-label="${text('provider')}"><option value="">${text('none')}</option></select><p id="provider-status" class="geo-status" aria-live="polite"></p></div></details></div><div class="geo-field">${field('geo',number('geo',4294967295))}
       <details id="geo-tools" class="geo-tools"><summary>${text('locationTools')}</summary>
        <div class="geo-toolbar" role="tablist" aria-label="${text('locationTools')}">${['browse','place','ip'].map((k,i)=>`<button id="geo-tab-${k}" type="button" role="tab" aria-controls="geo-panel-${k}" aria-selected="${i===0}" data-geo-tab="${k}" class="geo-tab ${i===0?'active':''}">${text('tool'+k)}</button>`).join('')}</div>
        <div id="geo-panel-browse" class="geo-panel" role="tabpanel" aria-labelledby="geo-tab-browse">
         <label class="label" for="country">${text('country')}</label><details class="search-picker" id="country-picker"><summary id="country-trigger"><span id="country-label">${text('chooseCountry')}</span></summary><div class="picker-popup"><label class="sr-only" for="country-filter">${text('countryFilter')}</label><input class="input" id="country-filter" type="search" placeholder="${text('countryFilterPlaceholder')}" autocomplete="off" disabled><select class="picker-list" id="country" size="7" aria-label="${text('country')}"><option value="">${text('chooseCountry')}</option></select><p id="country-status" class="geo-status" aria-live="polite"></p></div></details>
         <label class="label" for="city">${text('city')}</label><details class="search-picker" id="city-picker"><summary id="city-trigger"><span id="city-label">${text('chooseCity')}</span></summary><div class="picker-popup"><label class="sr-only" for="city-filter">${text('cityFilter')}</label><input class="input" id="city-filter" type="search" autocomplete="off" placeholder="${text('cityFilterPlaceholder')}" disabled><select class="picker-list" id="city" size="7" aria-label="${text('city')}" disabled><option value="">${text('chooseCity')}</option></select><button type="button" id="more-cities" data-variant="outline" class="btn btn-outline btn-sm" hidden>${text('loadMore')}</button></div></details>
         <div class="geo-actions"><button type="button" id="use-city" data-variant="outline" class="btn btn-outline btn-sm" disabled>${text('select')}</button></div><p id="city-status" class="geo-status" aria-live="polite"></p>
         <button type="button" id="geo-details" class="toggle-button geo-details-toggle" aria-pressed="false">${text('showDetailedPlaces')}</button><p class="geo-attribution"><a href="https://www.geonames.org/" target="_blank" rel="noopener noreferrer">GeoNames</a> · ${text('cityCoverage')}</p>
        </div>
        <div id="geo-panel-place" class="geo-panel" role="tabpanel" aria-labelledby="geo-tab-place" hidden><div class="lookup"><div class="input-group"><input class="input" id="place-query" placeholder="${text('placeQuery')}" aria-label="${text('placeQuery')}"><button type="button" id="place-search" data-variant="outline" class="btn btn-outline">${text('search')}</button></div><div id="places" aria-live="polite"></div></div></div>
        <div id="geo-panel-ip" class="geo-panel" role="tabpanel" aria-labelledby="geo-tab-ip" hidden><label class="label" for="lookup-ip">${text('lookupAddress')}</label><input class="input mono" id="lookup-ip" placeholder="192.0.2.10"><div class="geo-actions"><button type="button" id="geoip" data-variant="outline" class="btn btn-outline btn-sm">${text('lookup')}</button><button type="button" id="use-network-ips" data-variant="outline" class="btn btn-outline btn-sm">${text('ipLookup')}</button></div><div id="geoip-results" aria-live="polite"></div><p class="geo-attribution">${meta.geoipAttribution==='DB-IP'?'<a href="https://db-ip.com" target="_blank" rel="noopener noreferrer">IP Geolocation by DB-IP</a>':meta.geoipAttribution==='MaxMind'?'<a href="https://www.maxmind.com/" target="_blank" rel="noopener noreferrer">GeoIP data by MaxMind</a>':''}</p></div>
       </details></div></div></section>
      <section><div class="section-heading"><h2 class="section-label">${text('resources')}</h2>${help('intervalHelp')}</div><div class="field-grid">${field('architecture',select('architecture',optionList(['x86-64','ARM64','RISC-V 64',t('other')])))}${field('nodeType',select('nodeType',[{value:0,label:t('physical')},{value:1,label:t('vm')},{value:2,label:t('systemContainer')},{value:3,label:t('appContainer')},{value:7,label:t('other')}]))}</div><div class="field-grid resource-grid">${field('cpu',select('cpu',optionList(meta.cpu)))}${field('ram',select('ram',optionList(meta.ram)))}${field('storage',select('storage',optionList(meta.storage)))}</div><details class="recommendations"><summary>${text('recommendations')}</summary><div id="catalog-controls" class="catalog-controls"><label class="sr-only" for="plan-query">${text('planSearch')}</label><input id="plan-query" class="input" type="search" placeholder="${text('planSearch')}"><label class="sr-only" for="plan-family">${text('planFamily')}</label><select id="plan-family" class="select" aria-label="${text('planFamily')}"></select><button type="button" id="plan-match" class="toggle-button catalog-toggle" aria-pressed="false">${text('matchFields')}</button></div><div id="recommendations" class="plan-list"></div><div class="catalog-footer"><span id="plan-count" aria-live="polite"></span><button type="button" id="plan-more" class="btn btn-outline btn-sm" data-variant="outline">${text('loadMore')}</button><span id="catalog-date"></span></div></details></section>
      <section class="network-section"><h2 class="section-label">${text('network')}</h2><div class="field-grid">${field('ipv4','<input class="input mono" id="ipv4" placeholder="192.0.2.10" aria-label="IPv4" spellcheck="false">')}${field('ipv6','<input class="input mono" id="ipv6" placeholder="2001:db8::10" aria-label="IPv6" spellcheck="false">')}</div></section>
     </div>
    </form>
    <aside class="card result-panel">
     <div class="card-header"><h2 class="card-title">${text('output')}</h2></div>
     <div class="card-body">
      <div class="output-fields"><label class="label">${text('prefix')}<input class="input" form="encode-form" id="prefix" value="${escape(meta.prefix)}" maxlength="1" pattern="[a-z]" required></label><label class="label">${text('domain')}<input class="input" form="encode-form" id="domain" value="${escape(meta.domain)}" required spellcheck="false"></label></div>
      <label class="label" for="mode">${text('lengthMode')}</label><select class="select" form="encode-form" id="mode"><option value="compact">${text('compact')}</option><option value="padded">${text('padded')}</option></select>
      <div class="budget-block"><div class="section-heading"><h3>${text('used')}</h3>${help('capacityHelp')}</div><div class="budget"><span id="used-bits">—</span><span>/ 315 ${text('bits')}</span></div><meter id="budget-meter" min="0" max="315" value="0" aria-label="${text('used')}"></meter><div class="budget-meta"><span id="char-count">— / 63</span><span>${text('characters')}</span></div><span id="encoding-mode" class="sr-only">${text('pending')}</span></div>
      <button type="submit" form="encode-form" id="generate" class="btn w-100">${text('generate')}</button>
      <div class="hostname-block"><h3>${text('result')}</h3><p id="result-empty" class="text-secondary">${text('resultEmpty')}</p><output id="hostname" class="hostname" aria-live="polite"></output><p id="stale" class="text-warning" hidden>${text('stale')}</p><div id="result-actions" class="result-actions" hidden><button id="copy" data-variant="outline" class="btn btn-outline btn-sm">${text('copy')}</button><button id="export-result" data-variant="outline" class="btn btn-outline btn-sm">${text('export')}</button></div></div>
     </div>
    </aside>
   </div>
   <section id="decode-panel" class="card decode-panel" hidden><div class="card-header"><h2 class="card-title">${text('decodeInput')}</h2></div><div class="card-body"><form id="decode-form"><textarea class="input mono" id="decode-input" required maxlength="2048" placeholder="${text('decodePlaceholder')}" spellcheck="false"></textarea><button class="btn" type="submit">${text('decodeButton')}</button></form><div id="decoded" hidden><div class="section-heading"><h2 class="card-title">${text('decoded')}</h2><button id="edit-decoded" data-variant="outline" class="btn btn-outline btn-sm">${text('edit')}</button></div><dl id="decoded-values"></dl><p class="text-secondary">${text('decodeNote')}</p><button id="export-decoded" data-variant="outline" class="btn btn-outline btn-sm">${text('export')}</button></div></div></section>
   ${meta.demo?`<p class="geo-attribution"><a href="./runtime/NOTICE.txt">${text('demoLicenses')}</a></p>`:''}
  </main>
 </div>`;
 $<HTMLSelectElement>('#language').value=lang;$<HTMLSelectElement>('#theme').value=theme;$<HTMLSelectElement>('#mode').value=meta.mode;
 $('#language').addEventListener('change',()=>{const d=current(),s=settings();lang=$<HTMLSelectElement>('#language').value;localStorage.setItem('pnd.language',lang);mount();restore(d);$<HTMLInputElement>('#prefix').value=s.prefix;$<HTMLInputElement>('#domain').value=s.domain;$<HTMLSelectElement>('#mode').value=s.mode;pickerUI?.refresh();renderOutput();if(lastDecoded)renderDecoded(lastDecoded)});
 $('#theme').addEventListener('change',()=>{theme=$<HTMLSelectElement>('#theme').value;localStorage.setItem('pnd.theme',theme);applyTheme()});
 document.querySelectorAll<HTMLButtonElement>('[data-tab]').forEach(b=>b.addEventListener('click',()=>switchTab(b.dataset.tab!)));
 $('#encode-form').addEventListener('input',event=>{if((event.target as HTMLElement).closest('#geo-tools, #catalog-controls, .choice-filter'))return;if(lastHostname){outputStale=true;$('#stale').hidden=false}schedulePreview()});
 for(const key of ['prefix','domain','mode'])$('#'+key).addEventListener('input',invalidate);
 $('#new-id').addEventListener('click',async()=>{if(!confirm(t('confirmNewId')))return;try{nodeId=(await api<{nodeId:string}>('/api/node-id',{})).nodeId;$<HTMLInputElement>('#node-id').value=nodeId;invalidate()}catch(e){showError(e)}});
 $('#encode-form').addEventListener('submit',async ev=>{ev.preventDefault();clearError();const d=current();try{const result=await api<{descriptor:Descriptor;hostname:string;preview:Preview}>('/api/encode',{descriptor:d,...settings()});lastOutput=result.descriptor;lastHostname=result.hostname;outputStale=false;renderOutput();renderPreview(result.preview)}catch(e){showError(e)}});
 $('#copy').addEventListener('click',async()=>{const revision=outputRevision,hostname=lastHostname;try{if(navigator.clipboard&&window.isSecureContext)await navigator.clipboard.writeText(hostname);else{const el=document.createElement('textarea');el.value=hostname;document.body.append(el);el.select();if(!document.execCommand('copy'))throw Error(t('error'));el.remove()}if(revision===outputRevision)$('#copy').textContent=t('copied')}catch(e){showError(e)}});
 $('#export-input').addEventListener('click',()=>download(current()));$('#export-result').addEventListener('click',()=>lastOutput&&download(lastOutput));$('#export-decoded').addEventListener('click',()=>lastDecoded&&download(lastDecoded));
 $('#import').addEventListener('click',()=>$<HTMLInputElement>('#import-file').click());$('#import-file').addEventListener('change',async()=>{const f=$<HTMLInputElement>('#import-file').files?.[0];if(!f)return;try{if(f.size>16384)throw Error();const x=JSON.parse(await f.text()),d=x.descriptor||x;await api('/api/preview',{descriptor:d});restore(d);invalidate()}catch{showError(new Error(t('invalidImport')))}});
 $('#decode-form').addEventListener('submit',async ev=>{ev.preventDefault();clearError();++decodedLocationSequence;$('#decoded').hidden=true;lastDecoded=undefined;try{lastDecoded=await api<Descriptor>('/api/decode',{input:$<HTMLTextAreaElement>('#decode-input').value});renderDecoded(lastDecoded)}catch(e){showError(e)}});
 $('#edit-decoded').addEventListener('click',()=>{if(lastDecoded){restore(lastDecoded);invalidate();switchTab('encode')}});
 $('#place-search').addEventListener('click',searchPlaces);$('#geoip').addEventListener('click',()=>lookupIP(false));$('#use-network-ips').addEventListener('click',()=>lookupIP(true));
 $('#brand-home').addEventListener('click',()=>switchTab('encode'));
 setupGeoTools();
 if(session.required)$('#logout').addEventListener('click',()=>void logout());
 catalogUI=setupCatalog(meta,t,current,d=>{restore(d);invalidate()});
 pickerUI=setupPickers($('#app'),t);
 switchTab(tab);restore({nodeId});
}
function switchTab(next:string){tab=next;$('.page-heading h1').textContent=t(next==='encode'?'builder':'decode');for(const k of ['encode','decode'])$('#'+k+'-panel').hidden=k!==next;document.querySelectorAll<HTMLButtonElement>('[data-tab]').forEach(b=>{b.classList.toggle('active',b.dataset.tab===next);b.setAttribute('aria-current',b.dataset.tab===next?'page':'false')});clearError()}
function invalidate(){if(lastHostname){outputStale=true;$('#stale').hidden=false}schedulePreview()}
function schedulePreview(){clearTimeout(timer);const seq=++sequence;$('#encoding-mode').textContent=t('pending');timer=setTimeout(async()=>{try{const p=await api<Preview>('/api/preview',{descriptor:current(),...settings()});if(seq===sequence)renderPreview(p)}catch{if(seq===sequence){$('#used-bits').textContent='—';$('#char-count').textContent='— / 63';$('#encoding-mode').textContent=t('invalid_descriptor')}}},180)}
function renderPreview(p:Preview){$('#used-bits').textContent=String(p.bits);$<HTMLMeterElement>('#budget-meter').value=p.bits;$('#char-count').textContent=`${p.labelLength} / 63`;$('#encoding-mode').textContent=t(p.mode)}
function renderOutput(){if(!lastHostname)return;++outputRevision;$('#copy').textContent=t('copy');$('#hostname').textContent=lastHostname;$('#result-empty').hidden=true;$('#result-actions').hidden=false;$('#stale').hidden=!outputStale}
function valueLabel(k:string,v:unknown){if(k==='provider')return meta.providers.find(p=>p.id===v)?.name||String(v);if(k==='cpu'||k==='ram'||k==='storage')return meta[k][Number(v)];if(k==='architecture')return ['x86-64','ARM64','RISC-V 64',t('other')][Number(v)];if(k==='nodeType')return ({0:t('physical'),1:t('vm'),2:t('systemContainer'),3:t('appContainer'),7:t('other')} as Record<number,string>)[Number(v)];return String(v)}
function renderDecoded(d:Descriptor){
 const seq=++decodedLocationSequence;$('#decoded').hidden=false;$('#decoded-values').replaceChildren();
 for(const [k,v]of Object.entries(d)){
  const dt=document.createElement('dt'),dd=document.createElement('dd');dt.textContent=t(k);dd.textContent=valueLabel(k,v);$('#decoded-values').append(dt,dd);
  if(k==='geo'){for(const key of ['country','city']){const label=document.createElement('dt'),value=document.createElement('dd');label.textContent=t(key);value.id='decoded-'+key;value.textContent=t('loading');$('#decoded-values').append(label,value)}}
 }
 if(d.geo!==undefined)void resolveDecodedLocation(d.geo,seq,lang);
}
async function resolveDecodedLocation(id:number,seq:number,locale:string){
 let country='',city='';
 try{
  const result=await api<{found:boolean;country:{code:string;name:string};city:{name:string;names?:Record<string,string>}}>(`/api/location?id=${id}`);
  if(result.found){country=new Intl.DisplayNames([locale],{type:'region'}).of(result.country.code)||result.country.name;city=result.city.names?.[locale]||result.city.names?.en||result.city.name}
  else if(meta.placesEnabled){const place=await api<{countryCode?:string;countryName?:string;name?:string;fcl?:string}>(`/api/places?id=${id}`);country=place.countryCode&&/^[A-Z]{2}$/.test(place.countryCode)?new Intl.DisplayNames([locale],{type:'region'}).of(place.countryCode)||place.countryName||'':place.countryName||'';city=place.fcl==='P'?place.name||'':''}
  if(seq!==decodedLocationSequence)return;
  $('#decoded-country').textContent=country||t('locationNameUnavailable');$('#decoded-city').textContent=city||t('locationNameUnavailable');
 }catch{if(seq===decodedLocationSequence){$('#decoded-country').textContent=t('locationNameUnavailable');$('#decoded-city').textContent=t('locationNameUnavailable')}}
}
function download(d:Descriptor){const blob=new Blob([JSON.stringify({format:'pnd-export-v1',descriptor:d},null,2)],{type:'application/json'});const a=document.createElement('a');a.href=URL.createObjectURL(blob);a.download=`pnd-${d.nodeId}.json`;a.click();setTimeout(()=>URL.revokeObjectURL(a.href),1000)}
function recommend(){catalogUI?.refresh()}
async function searchPlaces(){clearError();const host=$('#places');host.replaceChildren();if(!meta.placesEnabled){host.textContent=t('places_unavailable');return}try{const rows=await api<{geonameId:number;name:string;countryName:string;adminName1:string;fcode:string}[]>('/api/places?q='+encodeURIComponent($<HTMLInputElement>('#place-query').value));for(const row of rows||[]){const b=document.createElement('button');b.type='button';b.dataset.variant='outline';b.className='btn btn-outline place-result';b.textContent=`${row.name} · ${row.adminName1} · ${row.countryName} · ${row.fcode} · ${row.geonameId}`;b.addEventListener('click',()=>{restore({...current(),geo:row.geonameId});invalidate()});host.append(b)}if(!host.childElementCount)host.textContent=t('noResults')}catch(e){showError(e)}}
function geoStatus(message:string){$('#city-status').textContent=message}
function applyLocation(id:number){restore({...current(),geo:id});invalidate()}
function syncPicker(kind:string){const select=$<HTMLSelectElement>('#'+kind);$('#'+kind+'-label').textContent=select.selectedOptions[0]?.textContent||t(kind==='country'?'chooseCountry':'chooseCity')}
function setupGeoTools(){
 const generation=mountGeneration;

 const enabled=availableGeoTools();
 $('#geo-tools').hidden=enabled.length===0;
 if(!enabled.length)return;
 const toolbar=$('.geo-toolbar');
 toolbar.addEventListener('click',event=>{const button=(event.target as HTMLElement).closest<HTMLButtonElement>('[data-geo-tab]');if(!button)return;switchGeoTool(button.dataset.geoTab!)});
 toolbar.addEventListener('keydown',event=>{const key=event.key;if(!['ArrowLeft','ArrowRight','Home','End'].includes(key))return;event.preventDefault();const buttons=Array.from(toolbar.querySelectorAll<HTMLButtonElement>('[data-geo-tab]')).filter(b=>!b.hidden);let i=buttons.indexOf(document.activeElement as HTMLButtonElement);i=key==='Home'?0:key==='End'?buttons.length-1:(i+(key==='ArrowRight'?1:buttons.length-1))%buttons.length;switchGeoTool(buttons[i].dataset.geoTab!);buttons[i].focus()});
 switchGeoTool(enabled[0]);if(!meta.countriesEnabled)return;let loaded=false;
 let countryRows:{code:string;name:string;english:string}[]=[];
 const fold=(value:string)=>value.normalize('NFD').replace(/[\u0300-\u036f]/g,'').toLocaleLowerCase(lang).trim();
 function filterCountries(){
  const select=$<HTMLSelectElement>('#country'),selected=select.value,q=fold($<HTMLInputElement>('#country-filter').value);
  const matches=countryRows.filter(c=>!q||[c.name,c.english,c.code].some(v=>fold(v).includes(q)));
  select.replaceChildren(new Option(t('chooseCountry'),''));
  const retained=countryRows.find(c=>c.code===selected);
  if(retained&&!matches.some(c=>c.code===selected)){const group=document.createElement('optgroup');group.label=t('selected');group.append(new Option(retained.name,retained.code));select.append(group)}
  for(const c of matches)select.add(new Option(c.name,c.code));
  select.value=selected;syncPicker('country');$('#country-status').textContent=q?(matches.length?`${matches.length} · ${t('country')}`:t('noResults')):'';
 }
 $('#country-filter').addEventListener('input',filterCountries);

 $('#geo-tools').addEventListener('toggle',async()=>{if(!$<HTMLDetailsElement>('#geo-tools').open||loaded)return;loaded=true;geoStatus(t('loading'));try{const rows=await api<{code:string;name:string;cities:number}[]>('/api/countries');if(generation!==mountGeneration)return;const names=new Intl.DisplayNames([lang],{type:'region'});countryRows=rows.map(r=>({code:r.code,name:names.of(r.code)||r.name,english:r.name})).sort((a,b)=>a.name.localeCompare(b.name,lang));$<HTMLInputElement>('#country-filter').disabled=false;filterCountries();geoStatus('')}catch(e){loaded=false;geoStatus(e instanceof Error?e.message:t('error'))}});
 $('#country').addEventListener('change',()=>{$<HTMLInputElement>('#city-filter').value='';void loadCities(false)});
 $('#city-filter').addEventListener('input',()=>{clearTimeout(cityTimer);++citySequence;$<HTMLButtonElement>('#use-city').disabled=true;cityTimer=setTimeout(()=>void loadCities(false),200)});
 $('#more-cities').addEventListener('click',()=>void loadCities(true));
 $('#geo-details').addEventListener('click',()=>{const b=$('#geo-details'),enabled=b.getAttribute('aria-pressed')!=='true';b.setAttribute('aria-pressed',String(enabled));b.textContent=t(enabled?'showMainPlaces':'showDetailedPlaces');void loadCities(false)});
 $('#city').addEventListener('change',()=>{$<HTMLButtonElement>('#use-city').disabled=$<HTMLSelectElement>('#city').value===''});
 $('#use-city').addEventListener('click',()=>{const row=cityRows.find(c=>String(c.id)===$<HTMLSelectElement>('#city').value);if(row)applyLocation(row.id)});
}
function availableGeoTools(){return ['browse','place','ip'].filter(k=>k==='browse'?meta.countriesEnabled:k==='place'?meta.placesEnabled:meta.geoipEnabled)}
function switchGeoTool(active:string){const enabled=availableGeoTools();if(!enabled.includes(active))active=enabled[0]||'';for(const name of ['browse','place','ip']){$('#geo-panel-'+name).hidden=name!==active;const b=$<HTMLButtonElement>('#geo-tab-'+name);b.hidden=!enabled.includes(name);b.classList.toggle('active',name===active);b.setAttribute('aria-selected',String(name===active));b.tabIndex=name===active?0:-1}}
async function loadCities(more:boolean){
 const seq=++citySequence,country=$<HTMLSelectElement>('#country').value,q=$<HTMLInputElement>('#city-filter').value;
 const city=$<HTMLSelectElement>('#city'),button=$<HTMLButtonElement>('#more-cities');
 $<HTMLInputElement>('#city-filter').disabled=!country;$<HTMLButtonElement>('#use-city').disabled=true;city.disabled=true;button.disabled=true;
 if(!more){cityRows=[];cityOffset=0;city.replaceChildren(new Option(t('chooseCity'),''));syncPicker('city');button.hidden=true}
 if(!country){geoStatus('');return}geoStatus(t('loading'));
 try{const result=await api<{cities:typeof cityRows;total:number;offset:number;more:boolean}>(`/api/cities?country=${encodeURIComponent(country)}&q=${encodeURIComponent(q)}&offset=${more?cityOffset:0}&details=${$('#geo-details').getAttribute('aria-pressed')==='true'}`);if(seq!==citySequence)return;
  cityRows.push(...result.cities);for(const c of result.cities){const name=c.names?.[lang]||c.names?.en||c.name,admin=c.adminNames?.[lang]||c.adminNames?.en||c.admin;city.add(new Option(`${name}${admin?' · '+admin:''} · ${c.id}`,String(c.id)))}
  syncPicker('city');cityOffset=result.offset+result.cities.length;city.disabled=cityRows.length===0;button.hidden=!result.more;button.disabled=false;geoStatus(result.total?`${cityRows.length} / ${result.total} · ${t('city')}`:t('noResults'));
 }catch(e){if(seq===citySequence){geoStatus(e instanceof Error?e.message:t('error'));button.disabled=false}}
}
async function lookupIP(useNetwork:boolean){
 clearError();const host=$('#geoip-results');host.replaceChildren();
 if(!meta.geoipEnabled){host.textContent=t('geoip_unavailable');return}
 const d=current(),ips=useNetwork?[d.ipv4,d.ipv6]:[$<HTMLInputElement>('#lookup-ip').value.trim()];
 if(!ips.some(Boolean)){host.textContent=t('public_ip_required');return}
 for(const ip of ips){if(!ip)continue;
  try{const row=await api<{ip:string;geo:number;name:string;country:string;admin:string;accuracyRadiusKm:number;candidates?:{id:number;name:string;admin:string;names?:Record<string,string>;adminNames?:Record<string,string>}[]}>('/api/geoip?ip='+encodeURIComponent(ip));
   const p=document.createElement('p');p.textContent=`${ip} · ${row.name||t('noCity')} · ${row.admin||''} · ${row.country||''}`;host.append(p);
   if(row.geo){const b=document.createElement('button');b.type='button';b.dataset.variant='outline';b.className='btn btn-outline btn-sm';b.textContent=t('select');b.addEventListener('click',()=>applyLocation(row.geo));host.append(b)}
   else if(row.candidates?.length){const note=document.createElement('p');note.className='muted';note.textContent=t('confirmGeoMatch');host.append(note);for(const c of row.candidates){const b=document.createElement('button');b.type='button';b.dataset.variant='outline';b.className='btn btn-outline place-result';b.textContent=`${c.names?.[lang]||c.names?.en||c.name} · ${c.adminNames?.[lang]||c.admin} · ${c.id}`;b.addEventListener('click',()=>applyLocation(c.id));host.append(b)}}
   else if(row.name){const note=document.createElement('p');note.className='muted';note.textContent=t('noGeoMatch');host.append(note)}
  }catch(e){const p=document.createElement('p');p.textContent=e instanceof Error?e.message:t('error');host.append(p)}
 }
 const note=document.createElement('p');note.className='muted';note.textContent=t('locationEstimate');host.append(note);
}
type AuthStatus='signedOut'|'signInError'|'signOutError'|'serviceError';
function authStatus(kind:AuthStatus){
 pickerUI?.destroy();clearTimeout(timer);clearTimeout(cityTimer);++citySequence;
 document.documentElement.lang=lang;
 const success=kind==='signedOut';
 const mark=success?'<path d="m6 12 4 4 8-8"/>':'<path d="m8 8 8 8M16 8l-8 8"/>';
 const action=kind==='signedOut'?'signInAgain':kind==='serviceError'?'reloadPage':kind==='signOutError'?'logout':'retrySignIn';
 $('#app').innerHTML=`<div class="auth-shell"><header class="app-header"><a class="brand-link" href="/"><strong>PND</strong>${session.brandName?`<span class="brand-name">${escape(session.brandName)}</span>`:''}</a></header><main class="auth-status"><section class="auth-status-panel" aria-labelledby="auth-title"><div class="auth-status-icon ${success?'is-success':'is-error'}" aria-hidden="true"><svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8" stroke-linecap="round" stroke-linejoin="round">${mark}</svg></div><h1 id="auth-title">${text(kind+'Title')}</h1><p class="auth-message" role="${success?'status':'alert'}">${text(kind+'Text')}</p><button type="button" class="btn" id="auth-action">${text(action)}</button></section></main></div>`;
 $('#auth-action').onclick=()=>{if(kind==='serviceError')location.reload();else if(kind==='signOutError')void logout();else location.assign('/auth/login')};
}
async function logout(){
 try{await api('/auth/logout',{});location.replace('/?auth=logged-out')}
 catch{authStatus('signOutError')}
}
async function start(){
 try{
  session=await api('/api/session');
  if(session.required&&!session.authenticated){
   const state=new URLSearchParams(location.search).get('auth');
   if(state==='logged-out')authStatus('signedOut');
   else if(state==='error')authStatus('signInError');
   else location.replace('/auth/login');
   return;
  }
  meta=await api('/api/meta');nodeId=(await api<{nodeId:string}>('/api/node-id',{})).nodeId;mount();
 }catch{authStatus('serviceError')}
}
void start();
