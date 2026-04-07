package server

import "net/http"

func (s *Server) dashboard(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write([]byte(dashboardHTML))
}

const dashboardHTML = `<!DOCTYPE html>
<html lang="en">
<head>
<meta charset="UTF-8">
<meta name="viewport" content="width=device-width,initial-scale=1.0">
<title>Headcount</title>
<link href="https://fonts.googleapis.com/css2?family=Libre+Baskerville:wght@400;700&family=JetBrains+Mono:wght@400;500;700&display=swap" rel="stylesheet">
<style>
:root{--bg:#1a1410;--bg2:#241e18;--bg3:#2e261e;--rust:#e8753a;--leather:#a0845c;--cream:#f0e6d3;--cd:#bfb5a3;--cm:#7a7060;--gold:#d4a843;--green:#4a9e5c;--red:#c94444;--mono:'JetBrains Mono',monospace;--serif:'Libre Baskerville',serif}
*{margin:0;padding:0;box-sizing:border-box}
body{background:var(--bg);color:var(--cream);font-family:var(--serif);line-height:1.6}
.header{padding:.9rem 1.5rem;border-bottom:1px solid var(--bg3);display:flex;justify-content:space-between;align-items:center;gap:1rem;flex-wrap:wrap}
.header h1{font-family:var(--mono);font-size:.9rem;letter-spacing:2px;display:flex;align-items:center;gap:.5rem}
.header h1 .live-dot{background:var(--green)}
.period-bar{display:flex;gap:.3rem;font-family:var(--mono);font-size:.65rem}
.period-btn{padding:.25rem .65rem;border:1px solid var(--bg3);background:var(--bg);color:var(--cm);cursor:pointer;font-family:var(--mono);transition:.15s}
.period-btn:hover{border-color:var(--leather)}
.period-btn.active{border-color:var(--rust);color:var(--rust)}
.content{padding:1.5rem;max-width:1100px;margin:0 auto}
.stats-row{display:grid;grid-template-columns:repeat(auto-fit,minmax(140px,1fr));gap:.8rem;margin-bottom:1.5rem}
.stat{background:var(--bg2);border:1px solid var(--bg3);padding:.9rem 1rem;text-align:center}
.stat-val{font-family:var(--mono);font-size:1.6rem;color:var(--cream);font-weight:700;display:flex;align-items:center;justify-content:center;gap:.4rem}
.stat-val .live-dot{background:var(--green)}
.stat-label{font-family:var(--mono);font-size:.55rem;color:var(--cm);text-transform:uppercase;letter-spacing:1px;margin-top:.3rem}
.chart-box{background:var(--bg2);border:1px solid var(--bg3);padding:1rem;margin-bottom:1rem}
.chart-title{font-family:var(--mono);font-size:.6rem;color:var(--leather);text-transform:uppercase;letter-spacing:1.5px;margin-bottom:.8rem}
.chart{height:140px;display:flex;align-items:flex-end;gap:3px}
.bar{background:var(--rust);min-width:4px;flex:1;position:relative;transition:opacity .15s;cursor:pointer}
.bar:hover{opacity:.7}
.bar-tip{display:none;position:absolute;bottom:100%;left:50%;transform:translateX(-50%);background:var(--bg);border:1px solid var(--bg3);padding:.25rem .5rem;font-family:var(--mono);font-size:.55rem;white-space:nowrap;z-index:10;color:var(--cream)}
.bar:hover .bar-tip{display:block}
.three-col{display:grid;grid-template-columns:1fr 1fr 1fr;gap:1rem}
.two-col{display:grid;grid-template-columns:1fr 1fr;gap:1rem}
@media(max-width:760px){.three-col,.two-col{grid-template-columns:1fr}}
.list-box{background:var(--bg2);border:1px solid var(--bg3);padding:1rem;margin-bottom:1rem}
.list-row{display:flex;justify-content:space-between;padding:.4rem 0;border-bottom:1px solid var(--bg3);font-family:var(--mono);font-size:.7rem;align-items:center;gap:.5rem}
.list-row:last-child{border:none}
.list-name{color:var(--cd);overflow:hidden;text-overflow:ellipsis;white-space:nowrap;flex:1;min-width:0}
.list-val{color:var(--cream);flex-shrink:0;font-weight:700}
.list-bar{height:3px;background:var(--rust);margin-top:.2rem;opacity:.4}
.live-dot{display:inline-block;width:8px;height:8px;border-radius:50%;animation:pulse 2s infinite}
@keyframes pulse{0%,100%{opacity:1}50%{opacity:.4}}
.empty{text-align:center;padding:1.5rem;color:var(--cm);font-style:italic;font-size:.75rem;font-family:var(--mono)}
.snippet-box{background:#0d0b09;border:1px solid var(--bg3);padding:.9rem 1rem;margin-top:1rem;font-family:var(--mono);font-size:.65rem;color:var(--cd);cursor:pointer;position:relative;line-height:1.5}
.snippet-box:hover{border-color:var(--leather)}
.copy-hint{position:absolute;top:.5rem;right:.6rem;font-size:.5rem;color:var(--cm);text-transform:uppercase;letter-spacing:1px}
.events-box{background:var(--bg2);border:1px solid var(--bg3);padding:1rem;margin-top:1rem;max-height:320px;overflow-y:auto}
.event-row{font-family:var(--mono);font-size:.6rem;padding:.35rem 0;border-bottom:1px solid var(--bg3);display:flex;gap:.6rem;align-items:center}
.event-row:last-child{border:none}
.event-time{color:var(--cm);width:60px;flex-shrink:0}
.event-name{color:var(--gold);width:60px;flex-shrink:0}
.event-page{color:var(--cream);overflow:hidden;text-overflow:ellipsis;white-space:nowrap;flex:1;min-width:0}
.event-meta{color:var(--cm);font-size:.55rem;flex-shrink:0}
</style>
</head>
<body>

<div class="header">
<h1 id="dash-title"><span class="live-dot"></span> HEADCOUNT</h1>
<div class="period-bar">
<button class="period-btn" onclick="setPeriod('today')" data-p="today">Today</button>
<button class="period-btn active" onclick="setPeriod('7d')" data-p="7d">7d</button>
<button class="period-btn" onclick="setPeriod('30d')" data-p="30d">30d</button>
<button class="period-btn" onclick="setPeriod('90d')" data-p="90d">90d</button>
</div>
</div>

<div class="content" id="main">
<div class="empty">Loading analytics&hellip;</div>
</div>

<script>
var API='/api';
var period='7d';
var liveTimer=null;

function setPeriod(p){
period=p;
document.querySelectorAll('.period-btn').forEach(function(b){
b.classList.toggle('active',b.getAttribute('data-p')===p);
});
load();
}

async function load(){
var q='?period='+period;
try{
var resps=await Promise.all([
fetch(API+'/stats'+q).then(function(r){return r.json()}),
fetch(API+'/pages'+q).then(function(r){return r.json()}),
fetch(API+'/referrers'+q).then(function(r){return r.json()}),
fetch(API+'/timeseries'+q).then(function(r){return r.json()}),
fetch(API+'/devices'+q).then(function(r){return r.json()}),
fetch(API+'/browsers'+q).then(function(r){return r.json()}),
fetch(API+'/countries'+q).then(function(r){return r.json()}),
fetch(API+'/events').then(function(r){return r.json()})
]);
render({
stats:resps[0],
pages:resps[1].pages||[],
refs:resps[2].referrers||[],
ts:resps[3].timeseries||[],
devs:resps[4]||{},
browsers:resps[5]||{},
countries:resps[6]||{},
events:resps[7].events||[]
});
}catch(e){
console.error('load failed',e);
document.getElementById('main').innerHTML='<div class="empty">Failed to load analytics</div>';
}
}

function render(d){
var m=document.getElementById('main');
var h='';

// Stats row
h+='<div class="stats-row">';
h+='<div class="stat"><div class="stat-val">'+fmt(d.stats.pageviews||0)+'</div><div class="stat-label">Pageviews</div></div>';
h+='<div class="stat"><div class="stat-val">'+fmt(d.stats.sessions||0)+'</div><div class="stat-label">Sessions</div></div>';
h+='<div class="stat"><div class="stat-val">'+(d.stats.bounce_rate||'0')+'%</div><div class="stat-label">Bounce Rate</div></div>';
h+='<div class="stat"><div class="stat-val"><span class="live-dot"></span><span id="live-count">'+fmt(d.stats.live_visitors||0)+'</span></div><div class="stat-label">Live Now</div></div>';
h+='</div>';

// Time series chart
h+='<div class="chart-box"><div class="chart-title">Pageviews over time</div>';
if(d.ts.length){
var max=1;
for(var i=0;i<d.ts.length;i++)if(d.ts[i].count>max)max=d.ts[i].count;
h+='<div class="chart">';
d.ts.forEach(function(t){
var pct=Math.max((t.count/max)*100,2);
h+='<div class="bar" style="height:'+pct+'%"><div class="bar-tip">'+esc(t.date)+': '+t.count+'</div></div>';
});
h+='</div>';
}else{
h+='<div class="empty">No data yet for this period</div>';
}
h+='</div>';

// Top pages + referrers (2 col)
h+='<div class="two-col">';
h+='<div class="list-box"><div class="chart-title">Top Pages</div>';
if(d.pages.length){
d.pages.slice(0,10).forEach(function(p){
h+='<div class="list-row"><span class="list-name">'+esc(p.page||'/')+'</span><span class="list-val">'+p.views+'</span></div>';
});
}else{
h+='<div class="empty">No pageviews yet</div>';
}
h+='</div>';

h+='<div class="list-box"><div class="chart-title">Top Referrers</div>';
if(d.refs.length){
d.refs.slice(0,10).forEach(function(r){
h+='<div class="list-row"><span class="list-name">'+esc(r.referrer)+'</span><span class="list-val">'+r.count+'</span></div>';
});
}else{
h+='<div class="empty">No referrers tracked</div>';
}
h+='</div>';
h+='</div>';

// Devices + Browsers + Countries (3 col)
h+='<div class="three-col">';
h+='<div class="list-box"><div class="chart-title">Devices</div>';
var devKeys=Object.keys(d.devs).sort(function(a,b){return d.devs[b]-d.devs[a]});
if(devKeys.length){
devKeys.forEach(function(k){
h+='<div class="list-row"><span class="list-name">'+esc(k)+'</span><span class="list-val">'+d.devs[k]+'</span></div>';
});
}else{
h+='<div class="empty">No data</div>';
}
h+='</div>';

h+='<div class="list-box"><div class="chart-title">Browsers</div>';
var brKeys=Object.keys(d.browsers).sort(function(a,b){return d.browsers[b]-d.browsers[a]});
if(brKeys.length){
brKeys.forEach(function(k){
h+='<div class="list-row"><span class="list-name">'+esc(k)+'</span><span class="list-val">'+d.browsers[k]+'</span></div>';
});
}else{
h+='<div class="empty">No data</div>';
}
h+='</div>';

h+='<div class="list-box"><div class="chart-title">Countries</div>';
var cKeys=Object.keys(d.countries).sort(function(a,b){return d.countries[b]-d.countries[a]});
if(cKeys.length){
cKeys.slice(0,10).forEach(function(k){
h+='<div class="list-row"><span class="list-name">'+esc(k)+'</span><span class="list-val">'+d.countries[k]+'</span></div>';
});
}else{
h+='<div class="empty">No country data</div>';
}
h+='</div>';
h+='</div>';

// Recent events log
h+='<div class="chart-title" style="margin-top:1.5rem">Recent Events</div>';
h+='<div class="events-box">';
if(d.events.length){
d.events.slice(0,30).forEach(function(e){
h+='<div class="event-row">';
h+='<span class="event-time">'+esc(fmtTime(e.created_at))+'</span>';
h+='<span class="event-name">'+esc(e.name||'event')+'</span>';
h+='<span class="event-page">'+esc(e.page||'-')+'</span>';
var meta=[];
if(e.country)meta.push(e.country);
if(e.device)meta.push(e.device);
if(meta.length)h+='<span class="event-meta">'+esc(meta.join(' / '))+'</span>';
h+='</div>';
});
}else{
h+='<div class="empty">No events yet</div>';
}
h+='</div>';

// Tracking snippet
h+='<div class="chart-title" style="margin-top:1.5rem">Tracking Snippet</div>';
h+='<div class="snippet-box" onclick="copySnippet(this)"><span class="copy-hint">click to copy</span><code id="snippet-code">'+esc(snippetCode())+'</code></div>';

m.innerHTML=h;
}

function snippetCode(){
return '<script>\nfetch("'+location.origin+'/api/event",{method:"POST",headers:{"Content-Type":"application/json"},body:JSON.stringify({page:location.pathname,referrer:document.referrer,name:"pageview"})});\n<\/script>';
}

function copySnippet(el){
var code=el.querySelector('code').textContent;
if(navigator.clipboard){
navigator.clipboard.writeText(code).then(function(){
var hint=el.querySelector('.copy-hint');
var orig=hint.textContent;
hint.textContent='copied!';
setTimeout(function(){hint.textContent=orig},1200);
});
}
}

function fmt(n){
if(n>=1000000)return(n/1000000).toFixed(1)+'M';
if(n>=1000)return(n/1000).toFixed(1)+'k';
return String(n);
}

function fmtTime(ts){
if(!ts)return'';
try{
var d=new Date(ts);
if(isNaN(d.getTime()))return ts;
return d.toLocaleTimeString('en-US',{hour:'2-digit',minute:'2-digit',second:'2-digit'});
}catch(e){return ts}
}

function esc(s){
if(s===undefined||s===null)return'';
var d=document.createElement('div');
d.textContent=String(s);
return d.innerHTML;
}

// ─── Personalization ──────────────────────────────────────────────

(function loadPersonalization(){
fetch('/api/config').then(function(r){return r.json()}).then(function(cfg){
if(!cfg||typeof cfg!=='object')return;
if(cfg.dashboard_title){
var h1=document.getElementById('dash-title');
if(h1)h1.innerHTML='<span class="live-dot"></span> '+esc(cfg.dashboard_title);
document.title=cfg.dashboard_title;
}
}).catch(function(){
}).finally(function(){
load();
// Live count refresh every 30s
liveTimer=setInterval(function(){
fetch(API+'/live').then(function(r){return r.json()}).then(function(d){
var el=document.getElementById('live-count');
if(el)el.textContent=fmt(d.live||0);
}).catch(function(){});
},30000);
});
})();
</script>
</body>
</html>`
