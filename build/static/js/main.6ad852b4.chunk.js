(this["webpackJsonpondemand-dash"]=this["webpackJsonpondemand-dash"]||[]).push([[0],{348:function(e,t,n){e.exports=n(485)},353:function(e,t,n){},485:function(e,t,n){"use strict";n.r(t);var r=n(0),a=n.n(r),o=n(15),c=n.n(o),s=(n(353),n(296)),i=n(564),l=n(563),m=n(565),u=n(201),d=n(571),E=n(566),h=n(556),p=n(569),v=n(568),f=n(572),g=n(567),w=(n(293),n(559),n(314),n(117),function(e){return r.createElement(d.a,e,r.createElement(E.a,null,r.createElement(h.a,{source:"fullname"}),r.createElement(h.a,{source:"mobile"}),r.createElement(h.a,{source:"username"}),r.createElement(p.a,{source:"is_provider"}),r.createElement(p.a,{source:"mobile_checked"}),r.createElement(p.a,{source:"is_active"}),r.createElement(u.a,{source:"description"}),r.createElement(h.a,{source:"latitude"}),r.createElement(h.a,{source:"longitude"})))}),k=function(e){return r.createElement(v.a,e,r.createElement(E.a,null,r.createElement(h.a,{disabled:!0,label:"Id",source:"id"}),r.createElement(h.a,{source:"fullname"}),r.createElement(u.a,{source:"description"}),r.createElement(p.a,{source:"is_provider"}),r.createElement(p.a,{source:"is_active"}),r.createElement(f.a,{helperText:"Set the service provider rating",source:"score",choices:[{id:1,name:"1"},{id:2,name:"2"},{id:3,name:"3"},{id:4,name:"4"},{id:5,name:"5"}]}),r.createElement(g.a,{source:"image"}),r.createElement(f.a,{source:"services",multiple:!0})))},b=n(159),P=n(295),S=n.n(P),j=n(315),y={login:function(e){var t=e.username,n=e.password,r=new Request("http://localhost:6662/admin/login",{method:"POST",body:JSON.stringify({username:t,password:n}),headers:new Headers({"Content-Type":"application/json"})});return fetch(r).then((function(e){if(e.status<200||e.status>=300)throw new Error(e.statusText);return e.json()})).then((function(e){var t=e.token;localStorage.setItem("token",t)}))},checkAuth:function(){return localStorage.getItem("token")?Promise.resolve():Promise.reject()},logout:function(){return localStorage.removeItem("token"),Promise.resolve()},getPermissions:function(){return localStorage.getItem("token")?Promise.resolve():Promise.reject()}},I=(Object(j.a)({direction:"rtl"}),Object(s.a)("https://ondemand.soluspay.net/admin")),O=Object(b.a)((function(){return S.a}),"ar"),_=function(){return r.createElement(i.a,{authProvider:y,dataProvider:I,i18nProvider:O},r.createElement(l.a,{name:"providers",list:m.a,edit:k,create:w}),r.createElement(l.a,{name:"users",list:m.a,edit:k,create:w}),r.createElement(l.a,{name:"count",list:m.a}),r.createElement(l.a,{name:"orders",list:m.a}))};Boolean("localhost"===window.location.hostname||"[::1]"===window.location.hostname||window.location.hostname.match(/^127(?:\.(?:25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)){3}$/));c.a.render(a.a.createElement(a.a.StrictMode,null,a.a.createElement(_,null)),document.getElementById("root")),"serviceWorker"in navigator&&navigator.serviceWorker.ready.then((function(e){e.unregister()})).catch((function(e){console.error(e.message)}))}},[[348,1,2]]]);
//# sourceMappingURL=main.6ad852b4.chunk.js.map