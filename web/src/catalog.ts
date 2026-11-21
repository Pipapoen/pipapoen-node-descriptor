export type CatalogDescriptor={nodeId:string;provider?:number;cpu?:number;ram?:number;storage?:number;architecture?:number;nodeType?:number};
export type Provider={id:number;name:string;aliases?:string[];website?:string};
export type Plan={provider:number;name:string;family:string;specs:string;cpu:number;ram:number;storage?:number;architecture?:number;nodeType?:number;source:string;checkedAt:string;kind:string};
type Metadata={providers:Provider[];plans:Plan[];cpu:string[];ram:string[];storage:string[];catalog:{checkedAt:string;plans:number}};

export function setupCatalog(meta:Metadata,t:(key:string)=>string,current:()=>CatalogDescriptor,apply:(descriptor:CatalogDescriptor)=>void){
 const $=<T extends HTMLElement>(id:string)=>document.getElementById(id) as T;
 const provider=$<HTMLSelectElement>('provider'),filter=$<HTMLInputElement>('provider-filter');
 const query=$<HTMLInputElement>('plan-query'),family=$<HTMLSelectElement>('plan-family'),match=$<HTMLButtonElement>('plan-match');
 let limit=12,lastProvider=provider.value;
 const label=(p:Provider)=>p.id===0?t('selfManaged'):p.id===65535?t('other'):p.name;
 const fold=(s:string)=>s.normalize('NFKC').toLocaleLowerCase().trim();
 const order=new Intl.Collator(document.documentElement.lang||'en',{sensitivity:'base',numeric:true});
 function filterProviders(){
  const saved=provider.value,retained=provider.selectedOptions[0]?.textContent||saved,q=fold(filter.value);
  const rows=meta.providers.filter(p=>!q||[label(p),p.name,String(p.id),...(p.aliases||[])].some(value=>fold(value).includes(q))).sort((a,b)=>order.compare(label(a),label(b))||a.id-b.id);
  provider.replaceChildren(new Option(t('none'),''));
  if(saved&&!rows.some(p=>String(p.id)===saved)){
   const group=document.createElement('optgroup');group.label=t('selected');
   const selected=meta.providers.find(p=>String(p.id)===saved);
   group.append(new Option(selected?label(selected):retained,saved));provider.append(group);
  }
  for(const special of [true,false]){
   const group=document.createElement('optgroup');group.label=t(special?'specialOptions':'providerCompanies');
   const members=rows.filter(p=>(p.id===0||p.id===65535)===special);
   if(special)members.sort((a,b)=>a.id-b.id);
   for(const p of members)group.append(new Option(label(p),String(p.id)));
   if(group.childElementCount)provider.append(group);
  }
  provider.value=saved;$('provider-label').textContent=provider.selectedOptions[0]?.textContent||t('none');
  $('provider-status').textContent=q?`${rows.length} · ${t('provider')}`:'';
 }
 const familyKey=(p:Plan)=>JSON.stringify([p.provider,p.family,p.kind==='plan']);
 function scoped(){const d=current();return meta.plans.filter(p=>d.provider===undefined||p.provider===d.provider)}
 function resourceOptions(){
  const d=current(),rows=scoped().filter(p=>(!family.value||familyKey(p)===family.value)&&(d.architecture===undefined||p.architecture===undefined||p.architecture===d.architecture));
  for(const key of ['cpu','ram','storage'] as const){
   const input=$<HTMLSelectElement>(key),saved=input.value;
   const compatible=rows.filter(p=>(key==='cpu'||d.cpu===undefined||p.cpu===d.cpu)&&(key!=='storage'||d.ram===undefined||p.ram===d.ram));
   const suggested=new Set(compatible.map(p=>p[key]).filter(v=>v!==undefined));
   input.replaceChildren(new Option(t('none'),''));
   for(const preferred of [true,false]){
    const group=document.createElement('optgroup');group.label=t(preferred?'recommendedOptions':'allOptions');
    meta[key].forEach((text,value)=>{if(suggested.has(value)===preferred)group.append(new Option(text,String(value)))});
    if(group.childElementCount)input.append(group);
   }
   input.value=saved;
  }
 }
 function render(){
  const d=current(),q=fold(query.value),matching=match.getAttribute('aria-pressed')==='true';
  let rows=scoped().filter(p=>(!family.value||familyKey(p)===family.value)&&(!q||fold([p.name,p.family,p.specs,meta.providers.find(v=>v.id===p.provider)?.name||''].join(' ')).includes(q)));
  const keys=['cpu','ram','storage','architecture','nodeType'] as const;
  if(matching)rows=rows.filter(p=>keys.every(key=>d[key]===undefined||p[key]===undefined||d[key]===p[key]));
  const score=(p:Plan)=>keys.reduce((n,key)=>n+(d[key]!==undefined&&p[key]===d[key]?1:0),0);
  rows.sort((a,b)=>score(b)-score(a));
  const host=$('recommendations');host.replaceChildren();
  for(const p of rows.slice(0,limit)){
   const row=document.createElement('div');row.className='plan-row';
   const description=document.createElement('div');description.className='plan-description';
   const name=document.createElement('strong');name.textContent=p.name;description.append(name);
   const context=document.createElement('span');context.className='plan-context';
   context.textContent=[d.provider===undefined?meta.providers.find(v=>v.id===p.provider)?.name:'',p.family,p.architecture===1?'ARM64':p.architecture===0?'x86-64':'',p.kind==='plan'?'':t('exampleConfiguration')].filter(Boolean).join(' · ');description.append(context);
   const specs=document.createElement('span');specs.className='plan-specs';specs.textContent=p.specs+(p.storage===undefined?' · '+t('separateStorage'):'');description.append(specs);row.append(description);
   const actions=document.createElement('div');actions.className='plan-actions';
   const source=document.createElement('a');source.href=p.source;source.target='_blank';source.rel='noopener noreferrer';source.textContent=t('planSource');source.title=`${t('checkedAt')} ${p.checkedAt}`;actions.append(source);
   const button=document.createElement('button');button.type='button';button.dataset.variant='outline';button.className='btn btn-outline btn-sm suggestion';button.textContent=t('apply');button.title=t('applyPlanHelp');button.dataset.plan=p.name;
   button.onclick=()=>{
    const next={...current(),provider:p.provider,cpu:p.cpu,ram:p.ram};
    for(const key of ['storage','architecture','nodeType'] as const)if(p[key]!==undefined)next[key]=p[key];
    apply(next);
   };
   actions.append(button);row.append(actions);host.append(row);
  }
  if(!rows.length){
   const empty=document.createElement('p');empty.className='muted';
   const directoryOnly=d.provider!==undefined&&scoped().length===0;
   empty.textContent=t(directoryOnly?'providerWithoutPlans':'resourceMismatch');host.append(empty);
   const website=meta.providers.find(p=>p.id===d.provider)?.website;
   if(directoryOnly&&website){const link=document.createElement('a');link.className='catalog-directory-link';link.href=website;link.target='_blank';link.rel='noopener noreferrer';link.textContent=t('providerWebsite');host.append(link)}
  }
  $('plan-count').textContent=`${Math.min(limit,rows.length)} / ${rows.length} · ${t('planCount')}`;
  $('plan-more').hidden=limit>=rows.length;
  $('catalog-date').textContent=`${t('checkedAt')} ${meta.catalog.checkedAt}`;
 }
 function refresh(){
  filterProviders();
  if(lastProvider!==provider.value){query.value='';family.value='';lastProvider=provider.value}
  const saved=family.value,showProvider=current().provider===undefined;
  const entries=Array.from(new Map(scoped().map(p=>[familyKey(p),p])).entries()).map(([value,p])=>{
   const owner=meta.providers.find(v=>v.id===p.provider);
   return {value,reference:p.kind!=='plan',text:[showProvider?(owner?label(owner):String(p.provider)):'',p.family].filter(Boolean).join(' \u00b7 ')};
  }).sort((a,b)=>order.compare(a.text,b.text));
  family.replaceChildren(new Option(t('allFamilies'),''));
  for(const reference of [false,true]){
   const group=document.createElement('optgroup');group.label=t(reference?'referenceConfigurations':'productFamilies');
   for(const entry of entries.filter(entry=>entry.reference===reference))group.append(new Option(entry.text,entry.value));
   if(group.childElementCount)family.append(group);
  }
  family.value=entries.some(entry=>entry.value===saved)?saved:'';
  limit=12;resourceOptions();render();
 }
 filter.oninput=filterProviders;
 provider.addEventListener('change',refresh);
 for(const key of ['cpu','ram','storage','architecture','nodeType'])$(key).addEventListener('change',refresh);
 query.addEventListener('input',()=>{limit=12;render()});
 family.addEventListener('change',()=>{limit=12;resourceOptions();render()});
 match.onclick=()=>{match.setAttribute('aria-pressed',String(match.getAttribute('aria-pressed')!=='true'));limit=12;render()};
 $('plan-more').onclick=()=>{limit+=12;render()};
 $('catalog-controls').addEventListener('keydown',event=>{if(event.key==='Enter'&&event.target===query)event.preventDefault()});
 refresh();
 return {refresh};
}
