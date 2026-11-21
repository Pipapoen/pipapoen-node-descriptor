import './style.css';
import examples from './examples.json';
import en from './locales/en.json';
import zh from './locales/zh-TW.json';

type Language='en'|'zh-TW';
type Mode='compact'|'padded';
let language:Language=localStorage.getItem('pnd.demo.language')==='zh-TW'?'zh-TW':'en';
let theme=localStorage.getItem('pnd.demo.theme')||'system';
if(!['light','dark','system'].includes(theme))theme='system';
let selected=1, mode:Mode='compact', generated=false, decoded=false, revision=0;
const $=<T extends HTMLElement>(id:string)=>document.getElementById(id) as T;
const t=(key:keyof typeof en)=>(language==='en'?en:zh)[key];
function applyTheme(){document.documentElement.dataset.theme=theme==='system'?(matchMedia('(prefers-color-scheme: dark)').matches?'dark':'light'):theme}
matchMedia('(prefers-color-scheme: dark)').addEventListener('change',applyTheme);
function details(){
 const d=examples[selected].descriptor;
 const rows=[[t('id'),d.nodeId]];
 if(d.geo)rows.push([t('locationLabel'),t('location')],[t('geo'),String(d.geo)]);
 if(d.provider!==undefined)rows.push([t('provider'),d.provider===0?t('selfManaged'):'DigitalOcean']);
 if(d.cpu!==undefined)rows.push([t('cpu'),d.cpu===1?'(1, 2]':'(3, 4]']);
 if(d.ram!==undefined)rows.push([t('ram'),d.ram===4?'(3, 4] GB':'(12, 16] GB']);
 if(d.storage!==undefined)rows.push([t('storage'),d.storage===2?'(64, 128] GB':'(512, 1024] GB']);
 if(d.ipv4)rows.push(['IPv4',d.ipv4]);
 return rows.map(([label,value])=>`<div><dt>${label}</dt><dd>${value}</dd></div>`).join('');
}
function render(){
 revision++;applyTheme();document.documentElement.lang=language;
 const result=examples[selected][mode];
 $('app').innerHTML=`<header><a class="brand" href="./">PND</a><nav><a href="https://github.com/Pipapoen/pipapoen-node-descriptor">${t('source')}</a><label class="sr-only" for="language">${t('language')}</label><select id="language"><option value="en">English</option><option value="zh-TW">繁體中文</option></select><label class="sr-only" for="theme">${t('theme')}</label><select id="theme">${['light','dark','system'].map(k=>`<option value="${k}">${t(k as keyof typeof en)}</option>`).join('')}</select></nav></header>
 <main><div class="heading"><div><p class="eyebrow">PND / DEMO</p><h1>${t('title')}</h1><p>${t('intro')}</p></div><a class="button secondary" href="https://github.com/Pipapoen/pipapoen-node-descriptor#deploy-with-docker-compose">${t('deploy')} <span aria-hidden="true">↗</span></a></div>
 <div class="workspace"><section aria-label="${t('node')}"><label for="node">${t('node')}</label><select id="node">${['identity','cloud','home'].map((key,i)=>`<option value="${i}">${t(key as keyof typeof en)}</option>`).join('')}</select><dl>${details()}</dl><label for="mode">${t('mode')}</label><select id="mode"><option value="compact">${t('compact')}</option><option value="padded">${t('padded')}</option></select><button id="generate">${t('generate')}</button></section>
 <section class="output" aria-live="polite"><p class="eyebrow">01 / ${t('result')}</p>${generated?`<code id="hostname">${result.hostname}</code><p class="count">${result.label.length} / 63 ${t('characters')}</p><div class="actions"><button class="secondary" id="copy">${t('copy')}</button><button class="secondary" id="decode">${t('decode')}</button></div>${decoded?`<div id="decoded"><p class="eyebrow">02 / ${t('details')}</p><dl>${details()}</dl></div>`:''}`:'<div class="empty" aria-hidden="true">x0… .example.org</div>'}</section></div>
 <footer><p>${t('notice')}</p><a href="https://www.geonames.org/">${t('attribution')}</a></footer></main>`;
 $<HTMLSelectElement>('language').value=language;$<HTMLSelectElement>('theme').value=theme;
 $<HTMLSelectElement>('node').value=String(selected);$<HTMLSelectElement>('mode').value=mode;
 $('language').onchange=()=>{language=$<HTMLSelectElement>('language').value as Language;localStorage.setItem('pnd.demo.language',language);render()};
 $('theme').onchange=()=>{theme=$<HTMLSelectElement>('theme').value;localStorage.setItem('pnd.demo.theme',theme);applyTheme()};
 $('node').onchange=()=>{selected=Number($<HTMLSelectElement>('node').value);generated=decoded=false;render()};
 $('mode').onchange=()=>{mode=$<HTMLSelectElement>('mode').value as Mode;generated=decoded=false;render()};
 $('generate').onclick=()=>{generated=true;decoded=false;render()};
 if(generated){
  $('decode').onclick=()=>{decoded=true;render()};
  $('copy').onclick=async()=>{const current=revision;try{await navigator.clipboard.writeText(result.hostname);if(current===revision)$('copy').textContent=t('copied')}catch{if(current===revision)$('copy').textContent=t('copyFailed')}};
 }
}
render();
