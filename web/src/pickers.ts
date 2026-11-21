// Shared searchable selection controls backed by native single-select listboxes.
export function setupPickers(root:HTMLElement,t:(key:string)=>string){
 const events=new AbortController();
 const controls:{select:HTMLSelectElement;picker:HTMLDetailsElement;refresh:()=>void;observer:MutationObserver}[]=[];
 for(const select of root.querySelectorAll<HTMLSelectElement>('select')){
  const id=select.id;
  const external=['provider','country','city'].includes(id);
  let picker=select.closest<HTMLDetailsElement>('details.search-picker');
  if(!picker){
   picker=document.createElement('details');picker.id=id+'-picker';picker.className='search-picker choice-picker';
   const summary=document.createElement('summary');summary.id=id+'-trigger';
   const label=document.createElement('span');label.id=id+'-label';summary.append(label);
   const popup=document.createElement('div');popup.className='picker-popup';
   const filter=document.createElement('input');filter.type='search';filter.id=id+'-filter';filter.className='input';filter.placeholder=t('filterOptions');filter.autocomplete='off';
   filter.setAttribute('aria-label',t('filterOptions'));
   const status=document.createElement('p');status.className='geo-status';status.id=id+'-status';status.setAttribute('aria-live','polite');
   select.replaceWith(picker);popup.append(filter,select,status);picker.append(summary,popup);
  }
  const wrapper=picker;
  wrapper.classList.add('choice-picker');
  const trigger=wrapper.querySelector<HTMLElement>('summary')!;
  const label=trigger.querySelector<HTMLElement>('span')!;
  const popup=wrapper.querySelector<HTMLElement>('.picker-popup')!;
  const filter=wrapper.querySelector<HTMLInputElement>('input[type=search]')!;
  const status=external?undefined:wrapper.querySelector<HTMLElement>('.geo-status');
  select.className='picker-list';select.size=7;
  filter.classList.add('choice-filter');
  const fieldLabel=root.querySelector<HTMLLabelElement>(`label[for="${id}"]`);
  const accessible=select.getAttribute('aria-label')||fieldLabel?.textContent||id;
  if(fieldLabel)fieldLabel.addEventListener('click',event=>{event.preventDefault();if(trigger.getAttribute('aria-disabled')!=='true'){trigger.focus();wrapper.open=true}},{signal:events.signal});
  trigger.setAttribute('aria-label',accessible);
  trigger.setAttribute('aria-controls',id+'-popup');trigger.setAttribute('aria-haspopup','listbox');
  popup.id=id+'-popup';
  filter.setAttribute('aria-controls',id);
  if(!external)filter.setAttribute('aria-label',`${t('filterOptions')}: ${accessible}`);
  const fold=(value:string)=>value.normalize('NFKC').toLocaleLowerCase().trim();
  let keyboardSelection=false;
  function place(){
   if(!wrapper.open)return;
   const rect=trigger.getBoundingClientRect();
   const width=Math.min(Math.max(rect.width,220),innerWidth-24);
   popup.style.width=width+'px';popup.style.right='auto';
   popup.style.left=Math.max(12-rect.left,Math.min(0,innerWidth-12-rect.left-width))+'px';
   const above=rect.top>innerHeight-rect.bottom&&rect.bottom+popup.offsetHeight+8>innerHeight;
   popup.style.top=above?'auto':'calc(100% + 5px)';popup.style.bottom=above?'calc(100% + 5px)':'auto';
  }
  function refresh(){
   if(!external){
    filter.disabled=select.disabled;
    const q=fold(filter.value);
    for(const option of select.options)option.hidden=Boolean(q&&option.value&&!fold(option.textContent+' '+option.value).includes(q));
    for(const group of select.querySelectorAll('optgroup'))group.hidden=![...group.children].some(option=>!(option as HTMLOptionElement).hidden);
    if(status)status.textContent=q&&![...select.options].some(o=>!o.hidden&&o.value)?t('noResults'):'';
   }
   label.textContent=select.selectedOptions[0]?.textContent||t('none');
   const disabled=select.disabled&&(!external||filter.disabled);
   trigger.setAttribute('aria-disabled',String(disabled));trigger.setAttribute('aria-expanded',String(wrapper.open));
   const rows=[...select.options].filter(o=>!o.hidden).length+select.querySelectorAll('optgroup:not([hidden])').length;
   select.style.height=Math.min(224,Math.max(64,rows*30))+'px';select.style.minHeight='0';
   place();
  }
  function close(focus=false){wrapper.open=false;trigger.setAttribute('aria-expanded','false');if(focus&&trigger.isConnected)trigger.focus()}
  trigger.addEventListener('click',event=>{if(trigger.getAttribute('aria-disabled')==='true')event.preventDefault()},{signal:events.signal});
  wrapper.addEventListener('toggle',()=>{
   trigger.setAttribute('aria-expanded',String(wrapper.open));
   if(wrapper.open){
    keyboardSelection=false;
    for(const control of controls)if(control.picker!==wrapper)control.picker.open=false;
    refresh();if(!filter.disabled){filter.focus();filter.select()}
   }
  },{signal:events.signal});
  wrapper.addEventListener('keydown',event=>{
   if(event.target===select)keyboardSelection=true;
   if(event.target===trigger&&trigger.getAttribute('aria-disabled')==='true'&&['ArrowDown','ArrowUp','Enter',' '].includes(event.key)){event.preventDefault();return}
   if(event.key==='Escape'){event.preventDefault();event.stopPropagation();close(true)}
   else if(event.key==='ArrowDown'&&(event.target===filter||event.target===trigger)){
    event.preventDefault();if(!wrapper.open)wrapper.open=true;else if(!select.disabled)select.focus();
   }else if(event.key==='Enter'&&event.target===select){event.preventDefault();close(true)}
   else if(event.key==='Enter'&&event.target===filter){event.preventDefault();if(!select.disabled)select.focus()}
  },{signal:events.signal});
  wrapper.addEventListener('focusout',event=>{if(event.relatedTarget&&!wrapper.contains(event.relatedTarget as Node))close()},{signal:events.signal});
  select.addEventListener('pointerdown',()=>{keyboardSelection=false},{signal:events.signal});
  select.addEventListener('keyup',()=>{keyboardSelection=false},{signal:events.signal});
  // Commit before dependent fields rebuild options or a locale change remounts the form.
  select.addEventListener('change',()=>{if(!keyboardSelection)close(true);refresh()},{capture:true,signal:events.signal});
  select.addEventListener('click',event=>{if(event.target instanceof HTMLOptionElement)close(true)},{signal:events.signal});
  filter.addEventListener('input',refresh,{signal:events.signal});
  const observer=new MutationObserver(refresh);observer.observe(select,{childList:true,subtree:true,attributes:true,attributeFilter:['disabled','label']});
  controls.push({select,picker:wrapper,refresh,observer});refresh();
 }
 document.addEventListener('pointerdown',event=>{for(const control of controls)if(!control.picker.contains(event.target as Node))control.picker.open=false},{signal:events.signal});
 window.addEventListener('resize',()=>{for(const control of controls)if(control.picker.open)control.refresh()},{signal:events.signal});
 return {refresh:()=>controls.forEach(control=>control.refresh()),destroy:()=>{events.abort();controls.forEach(control=>control.observer.disconnect())}};
}
