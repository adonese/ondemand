(this["webpackJsonpondemand-dash"]=this["webpackJsonpondemand-dash"]||[]).push([[0],{357:function(e,a,t){e.exports=t(495)},362:function(e,a,t){},495:function(e,a,t){"use strict";t.r(a);var n=t(0),r=t.n(n),i=t(15),c=t.n(i),o=(t(362),t(302)),m=t(574),s=t(573),l=t(575),d=t(207),u=t(583),E=t(576),h=t(565),p=t(580),v=t(578),f=t(568),g=t(577),b=t(579),w=(t(299),t(120)),_=(t(569),t(322),function(e){return n.createElement(u.a,e,n.createElement(E.a,null,n.createElement(h.a,{source:"fullname"}),n.createElement(h.a,{source:"mobile"}),n.createElement(h.a,{source:"username"}),n.createElement(p.a,{source:"is_provider"}),n.createElement(p.a,{source:"mobile_checked"}),n.createElement(p.a,{source:"is_active"}),n.createElement(d.a,{source:"description"}),n.createElement(h.a,{source:"latitude"}),n.createElement(h.a,{source:"longitude"})))}),k=[{id:40,name:"vehicle"},{id:39,name:"ac"},{id:38,name:"flies"},{id:37,name:"disinfectant"},{id:36,name:"sink"},{id:35,name:"electricity"},{id:34,name:"delivery"},{id:33,name:"carpenter"},{id:32,name:"colors"},{id:31,name:"ceramics"},{id:30,name:"smith"},{id:29,name:"alumnuim"},{id:28,name:"kitchen"},{id:27,name:"electronics"},{id:26,name:"mobile_fix"},{id:25,name:"cctv"},{id:24,name:"parties"},{id:23,name:"safari"},{id:22,name:"sweets"},{id:21,name:"family_biz"},{id:20,name:"taxi"},{id:19,name:"wiw"},{id:18,name:"barber"},{id:17,name:"wash"},{id:16,name:"coop"},{id:15,name:"vegan"},{id:14,name:"meat"},{id:13,name:"coffee"},{id:12,name:"cook"},{id:11,name:"pencils"},{id:10,name:"teachers"},{id:9,name:"students"},{id:8,name:"car_ele"},{id:7,name:"car_mech"},{id:6,name:"batteries"},{id:5,name:"car_mech"},{id:4,name:"long_vehicle"},{id:3,name:"soccer"},{id:2,name:"marriage"},{id:1,name:"car_wash"}],y=function(e){return n.createElement(v.a,e,n.createElement(E.a,null,n.createElement(h.a,{disabled:!0,label:"Id",source:"id"}),n.createElement(h.a,{source:"fullname"}),n.createElement(d.a,{source:"description"}),n.createElement(p.a,{source:"is_provider"}),n.createElement(p.a,{source:"is_active"}),n.createElement(p.a,{source:"mobile_checked"}),n.createElement(p.a,{source:"is_disabled",label:"Disable service provider?"}),n.createElement(h.a,{source:"latitude"}),n.createElement(h.a,{source:"longitude"}),n.createElement(f.a,{multiple:!0,helperText:"Set the service provider rating",source:"score",choices:[{id:1,name:"1"},{id:2,name:"2"},{id:3,name:"3"},{id:4,name:"4"},{id:5,name:"5"}]}),n.createElement(g.a,{source:"image"}),n.createElement(b.a,{source:"services",choices:k,translateChoice:!1})))},P=function(e){return n.createElement(v.a,e,n.createElement(E.a,null,n.createElement(w.a,{disabled:!0,label:"Id",source:"id"}),n.createElement(w.a,{source:"user_id"}),n.createElement(w.a,{source:"provider_id"}),n.createElement(w.a,{source:"uuid"}),n.createElement(f.a,{multiple:!0,helperText:"Service ID",source:"category",choices:k})))},S=t(163),I=t(301),j=t.n(I),x=t(323),O={login:function(e){var a=e.username,t=e.password,n=new Request("https://ondemand.soluspay.net/admin/login",{method:"POST",body:JSON.stringify({username:a,password:t}),headers:new Headers({"Content-Type":"application/json"})});return fetch(n).then((function(e){if(e.status<200||e.status>=300)throw new Error(e.statusText);return e.json()})).then((function(e){var a=e.token;localStorage.setItem("token",a)}))},checkAuth:function(){return localStorage.getItem("token")?Promise.resolve():Promise.reject()},logout:function(){return localStorage.removeItem("token"),Promise.resolve()},getPermissions:function(){return localStorage.getItem("token")?Promise.resolve():Promise.reject()}},T=t(582),J=(Object(x.a)({direction:"rtl"}),Object(o.a)("https://ondemand.soluspay.net/admin")),B=Object(S.a)((function(){return j.a}),"ar"),C=function(e){return n.createElement(T.a,e,n.createElement(h.a,{label:"Search",source:"name",alwaysOn:!0}))},D=function(){return n.createElement(m.a,{authProvider:O,dataProvider:J,i18nProvider:B},n.createElement(s.a,{name:"providers",list:l.a,edit:y,create:_}),n.createElement(s.a,{name:"users",list:l.a,edit:y,create:_}),n.createElement(s.a,{name:"count",list:l.a}),n.createElement(s.a,{name:"orders",list:l.a,edit:P,filters:n.createElement(C,null)}))};Boolean("localhost"===window.location.hostname||"[::1]"===window.location.hostname||window.location.hostname.match(/^127(?:\.(?:25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)){3}$/));c.a.render(r.a.createElement(r.a.StrictMode,null,r.a.createElement(D,null)),document.getElementById("root")),"serviceWorker"in navigator&&navigator.serviceWorker.ready.then((function(e){e.unregister()})).catch((function(e){console.error(e.message)}))}},[[357,1,2]]]);
//# sourceMappingURL=main.afed1e51.chunk.js.map