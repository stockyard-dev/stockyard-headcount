package server

import "net/http"

func (s *Server) dashboard(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")
	w.Write([]byte(dashboardHTML))
}

const dashboardHTML = `<!DOCTYPE html>
<html lang="en">
<head>
<meta charset="UTF-8"><meta name="viewport" content="width=device-width,initial-scale=1.0">
<title>Headcount</title>
<style>
:root{--bg:#1a1410;--bg2:#241e18;--bg3:#2e261e;--rust:#e8753a;--leather:#a0845c;--cream:#f0e6d3;--cd:#bfb5a3;--cm:#7a7060;--gold:#d4a843;--green:#4a9e5c;--mono:'JetBrains Mono',monospace;--serif:'Libre Baskerville',serif}
*{margin:0;padding:0;box-sizing:border-box}body{background:var(--bg);color:var(--cream);font-family:var(--serif);line-height:1.6}
.header{padding:1rem 1.5rem;border-bottom:1px solid var(--bg3);display:flex;justify-content:space-between;align-items:center}
.header h1{font-family:var(--mono);font-size:.9rem;letter-spacing:2px}
.period-bar{display:flex;gap:.3rem;font-family:var(--mono);font-size:.65rem}
.period-btn{padding:.2rem .6rem;border:1px solid var(--bg3);background:var(--bg);color:var(--cm);cursor:pointer}
.period-btn:hover{border-color:var(--leather)}.period-btn.active{border-color:var(--rust);color:var(--rust)}
.content{padding:1.5rem;max-width:1000px;margin:0 auto}
.stats-row{display:grid;grid-template-columns:repeat(auto-fit,minmax(130px,1fr));gap:.8rem;margin-bottom:1.5rem}
.stat{background:var(--bg2);border:1px solid var(--bg3);padding:.8rem 1rem;text-align:center}
.stat-val{font-family:var(--mono);font-size:1.6rem;color:var(--cream)}
.stat-label{font-family:var(--mono);font-size:.6rem;color:var(--cm);text-transform:uppercase;letter-spacing:1px;margin-top:.2rem}
.chart-box{background:var(--bg2);border:1px solid var(--bg3);padding:1rem;margin-bottom:1rem}
.chart-title{font-family:var(--mono);font-size:.65rem;color:var(--leather);text-transform:uppercase;letter-spacing:1px;margin-bottom:.8rem}
.chart{height:120px;display:flex;align-items:flex-end;gap:2px}
.bar{background:var(--rust);min-width:4px;flex:1;border-radius:1px 1px 0 0;position:relative;transition:opacity .15s;cursor:pointer}
.bar:hover{opacity:.7}
.bar-tip{display:none;position:absolute;bottom:100%;left:50%;transform:translateX(-50%);background:var(--bg);border:1px solid var(--bg3);padding:.2rem .4rem;font-family:var(--mono);font-size:.6rem;white-space:nowrap;z-index:10}
.bar:hover .bar-tip{display:block}
.two-col{display:grid;grid-template-columns:1fr 1fr;gap:1rem}
@media(max-width:600px){.two-col{grid-template-columns:1fr}}
.list-box{background:var(--bg2);border:1px solid var(--bg3);padding:1rem;margin-bottom:1rem}
.list-row{display:flex;justify-content:space-between;padding:.35rem 0;border-bottom:1px solid var(--bg3);font-family:var(--mono);font-size:.75rem}
.list-row:last-child{border:none}
.list-name{color:var(--cd);overflow:hidden;text-overflow:ellipsis;white-space:nowrap;max-width:70%}
.list-val{color:var(--cream);flex-shrink:0}
.live-dot{display:inline-block;width:8px;height:8px;border-radius:50%;background:var(--green);margin-right:.3rem;animation:pulse 2s infinite}
@keyframes pulse{0%,100%{opacity:1}50%{opacity:.4}}
.empty{text-align:center;padding:2rem;color:var(--cm);font-style:italic;font-size:.85rem}
.snippet-box{background:#0d0b09;border:1px solid var(--bg3);padding:.8rem 1rem;margin-top:1rem;font-family:var(--mono);font-size:.7rem;color:var(--cd);cursor:pointer;position:relative}
.snippet-box:hover{border-color:var(--leather)}
.copy-hint{position:absolute;top:.5rem;right:.5rem;font-size:.55rem;color:var(--cm)}
</style>
</head>
<body>
<div class="header">
  <h1><span class="live-dot"></span>HEADCOUNT</h1>
  <div class="period-bar">
    <button class="period-btn" onclick="setPeriod('today')">Today</button>
    <button class="period-btn active" onclick="setPeriod('7d')">7d</button>
    <button class="period-btn" onclick="setPeriod('30d')">30d</button>
    <button class="period-btn" onclick="setPeriod('90d')">90d</button>
  </div>
</div>
<div class="content" id="main">
  <div class="empty">Loading analytics...</div>
</div>

<script>
const API='/api';
let period='7d';

function setPeriod(p){
  period=p;
  document.querySelectorAll('.period-btn').forEach(b=>b.classList.toggle('active',b.textContent.toLowerCase()===p));
  load();
}

async function load(){
  const q='?period='+period;
  const[stats,pages,refs,ts,devs,browsers]=await Promise.all([
    fetch(API+'/stats'+q).then(r=>r.json()),
    fetch(API+'/pages'+q).then(r=>r.json()),
    fetch(API+'/referrers'+q).then(r=>r.json()),
    fetch(API+'/timeseries'+q).then(r=>r.json()),
    fetch(API+'/devices'+q).then(r=>r.json()),
    fetch(API+'/browsers'+q).then(r=>r.json()),
  ]);
  render(stats,pages.pages||[],refs.referrers||[],ts.timeseries||[],devs,browsers);
}

function render(stats,pages,refs,ts,devs,browsers){
  const m=document.getElementById('main');
  let h='';

  // Stats row
  h+='<div class="stats-row">';
  h+='<div class="stat"><div class="stat-val">'+fmt(stats.pageviews||0)+'</div><div class="stat-label">Pageviews</div></div>';
  h+='<div class="stat"><div class="stat-val">'+fmt(stats.sessions||0)+'</div><div class="stat-label">Sessions</div></div>';
  h+='<div class="stat"><div class="stat-val">'+(stats.bounce_rate||'0')+'%</div><div class="stat-label">Bounce Rate</div></div>';
  h+='<div class="stat"><div class="stat-val"><span class="live-dot"></span>'+fmt(stats.live_visitors||0)+'</div><div class="stat-label">Live Now</div></div>';
  h+='</div>';

  // Chart
  h+='<div class="chart-box"><div class="chart-title">Pageviews over time</div>';
  if(ts.length){
    const max=Math.max(...ts.map(t=>t.count),1);
    h+='<div class="chart">';
    ts.forEach(t=>{
      const pct=Math.max((t.count/max)*100,2);
      h+='<div class="bar" style="height:'+pct+'%"><div class="bar-tip">'+t.date+': '+t.count+'</div></div>';
    });
    h+='</div>';
  } else {h+='<div class="empty">No data yet for this period</div>';}
  h+='</div>';

  // Two columns
  h+='<div class="two-col">';

  // Top pages
  h+='<div class="list-box"><div class="chart-title">Top Pages</div>';
  if(pages.length){pages.slice(0,10).forEach(p=>{h+='<div class="list-row"><span class="list-name">'+esc(p.page||'/')+'</span><span class="list-val">'+p.views+'</span></div>';});}
  else{h+='<div class="empty">No pageviews yet</div>';}
  h+='</div>';

  // Referrers
  h+='<div class="list-box"><div class="chart-title">Top Referrers</div>';
  if(refs.length){refs.slice(0,10).forEach(r=>{h+='<div class="list-row"><span class="list-name">'+esc(r.referrer)+'</span><span class="list-val">'+r.count+'</span></div>';});}
  else{h+='<div class="empty">No referrers tracked</div>';}
  h+='</div>';

  h+='</div>';

  // Devices + Browsers
  h+='<div class="two-col">';
  h+='<div class="list-box"><div class="chart-title">Devices</div>';
  Object.entries(devs).sort((a,b)=>b[1]-a[1]).forEach(([k,v])=>{h+='<div class="list-row"><span class="list-name">'+esc(k)+'</span><span class="list-val">'+v+'</span></div>';});
  if(!Object.keys(devs).length)h+='<div class="empty">No data</div>';
  h+='</div>';

  h+='<div class="list-box"><div class="chart-title">Browsers</div>';
  Object.entries(browsers).sort((a,b)=>b[1]-a[1]).forEach(([k,v])=>{h+='<div class="list-row"><span class="list-name">'+esc(k)+'</span><span class="list-val">'+v+'</span></div>';});
  if(!Object.keys(browsers).length)h+='<div class="empty">No data</div>';
  h+='</div>';
  h+='</div>';

  // Tracking snippet
  h+='<div class="chart-title" style="margin-top:1.5rem">Tracking Snippet</div>';
  h+='<div class="snippet-box" onclick="navigator.clipboard.writeText(this.querySelector(\'code\').textContent)"><span class="copy-hint">click to copy</span><code>&lt;script&gt;\nfetch("'+location.origin+'/api/event",{method:"POST",headers:{"Content-Type":"application/json"},body:JSON.stringify({page:location.pathname,referrer:document.referrer,name:"pageview"})});\n&lt;/script&gt;</code></div>';

  m.innerHTML=h;
}

function fmt(n){return n>=1000?(n/1000).toFixed(1)+'k':n.toString();}
function esc(s){if(!s)return'';const d=document.createElement('div');d.textContent=s;return d.innerHTML;}

load();
setInterval(()=>{fetch(API+'/live').then(r=>r.json()).then(d=>{const el=document.querySelector('.stat-val .live-dot');if(el)el.parentElement.childNodes[1].textContent=d.live||0;})},30000);
</script>
</body>
</html>`
