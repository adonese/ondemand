(this["webpackJsonpondemand-dash"]=this["webpackJsonpondemand-dash"]||[]).push([[0],{344:function(e,t,n){e.exports=n(481)},349:function(e,t,n){},481:function(e,t,n){"use strict";n.r(t);var r=n(0),a=n.n(r),o=n(15),c=n.n(o),s=(n(349),n(292)),i=n(562),l=n(561),u=n(563),m=n(199),d=n(569),h=n(564),E=n(313),f=n(555),p=n(566),v=n(565),g=(n(289),n(557),n(310),n(116),function(e){return r.createElement(d.a,e,r.createElement(h.a,null,r.createElement(E.a,{source:"fullname"}),r.createElement(E.a,{source:"mobile"}),r.createElement(f.a,{source:"password"}),r.createElement(E.a,{source:"username"}),r.createElement(p.a,{source:"is_provider"}),r.createElement(p.a,{source:"is_active"}),r.createElement(m.a,{source:"description"})))}),w=function(e){return r.createElement(v.a,e,r.createElement(h.a,null,r.createElement(E.a,{disabled:!0,label:"Id",source:"id"}),r.createElement(E.a,{source:"fullname"}),r.createElement(m.a,{source:"description"}),r.createElement(p.a,{source:"is_active"})))},k=n(156),P=n(291),b=n.n(P),j=n(311),S={login:function(e){var t=e.username,n=e.password,r=new Request("http://localhost:6662/admin/login",{method:"POST",body:JSON.stringify({username:t,password:n}),headers:new Headers({"Content-Type":"application/json"})});return fetch(r).then((function(e){if(e.status<200||e.status>=300)throw new Error(e.statusText);return e.json()})).then((function(e){var t=e.token;localStorage.setItem("token",t)}))},checkAuth:function(){return localStorage.getItem("token")?Promise.resolve():Promise.reject()},logout:function(){return localStorage.removeItem("token"),Promise.resolve()},getPermissions:function(){return localStorage.getItem("token")?Promise.resolve():Promise.reject()}},I=(Object(j.a)({direction:"rtl"}),Object(s.a)("http://localhost:6662/admin")),y=Object(k.a)((function(){return b.a}),"ar"),O=function(){return r.createElement(i.a,{authProvider:S,dataProvider:I,i18nProvider:y},r.createElement(l.a,{name:"providers",list:u.a,edit:w,create:g}),r.createElement(l.a,{name:"count",list:u.a}),r.createElement(l.a,{name:"orders",list:u.a}))};Boolean("localhost"===window.location.hostname||"[::1]"===window.location.hostname||window.location.hostname.match(/^127(?:\.(?:25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)){3}$/));c.a.render(a.a.createElement(a.a.StrictMode,null,a.a.createElement(O,null)),document.getElementById("root")),"serviceWorker"in navigator&&navigator.serviceWorker.ready.then((function(e){e.unregister()})).catch((function(e){console.error(e.message)}))}},[[344,1,2]]]);
//# sourceMappingURL=main.498da11b.chunk.js.map