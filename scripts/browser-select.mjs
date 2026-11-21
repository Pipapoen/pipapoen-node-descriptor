// Exercise the shared selection control through its visible popup.
export async function pick(page,id,value){
 const wrapper=page.locator('#'+id+'-picker');
 if(!await wrapper.evaluate(element=>element.open))await page.locator('#'+id+'-trigger').click();
 const select=page.locator('#'+id);
 const available=await select.evaluate((element,value)=>[...element.options].some(option=>option.value===value&&!option.hidden),value);
 if(!available)await page.locator('#'+id+'-filter').fill('');
 await select.selectOption(value);
 if(await wrapper.evaluate(element=>element.open))await select.press('Enter');
}
