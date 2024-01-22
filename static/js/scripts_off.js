var buttons = document.getElementsByClassName("button");
for(let i = 0; i < buttons.length; i++){
	if(buttons[i].getAttribute("onclick") != ""){
		buttons[i].addEventListener('keyup', (event) => {
			if(event.keyCode == 13 || event.keyCode == 32)
				eval(buttons[i].getAttribute("onclick"));
		});
	}
}

onload = () => {
	setStyles();
}

window.onresize = () => {
	setStyles();
}

function setStyles(){
	if(window.innerWidth >= window.innerHeight){ //横画面の場合
		document.body.style.padding = "0 35%";
	}
	else{ //縦画面の場合
		document.body.style.padding = "0";
	}
}

function roommake_next(){
	var error = false;
	if(document.fm.playercount.value < 3){
		document.getElementById("error_playercount").style.display = "inline-block";
		error = true;
	}
	else document.getElementById("error_playercount").style.display = "none";

	if(document.fm.wordwolfcount.value < 1){
		document.getElementById("error_wordwolfcount").style.display = "inline-block";
		error = true;
	}
	else document.getElementById("error_wordwolfcount").style.display = "none";

	if(error){
		alert("入力内容に誤りがあります。");
		return;
	}
	
	sessionStorage.setItem("playercount", document.fm.playercount.value);
	sessionStorage.setItem("wordwolfcount", document.fm.wordwolfcount.value);
	sessionStorage.setItem("talkcategory", document.fm.talkcategory.value);
	var talktime = 0;
	var talktime_str = document.fm.talktime.value.split(":");
	talktime += (talktime_str[0] - 0) * 3600;
	talktime += (talktime_str[1] - 0) * 60;
	if(talktime_str.length == 3)
		talktime += (talktime_str[2] - 0);
	sessionStorage.setItem("talktime", talktime);
	fetch('/r/allquestions', {
		method: "GET"
	}).then((res) => {
		return res.json();
	}).then((obj) => {
		if(obj.result_type == 0){
			var q = new Array();
			for(let i = 0; i < Object.keys(obj.question).length; i++){
				if(obj.question[i].category == document.fm.talkcategory.value)
					q.push(obj.question[i]);
			}
			var rnd = Math.floor(Math.random() * Math.floor(q.length));
			var map = makeStr('0', (document.fm.playercount.value - 0) - (document.fm.wordwolfcount.value - 0)) + makeStr('1', (document.fm.wordwolfcount.value - 0));
			map = map.split("").sort(() => Math.random() - 0.5);
			console.log("map", map);

			sessionStorage.setItem("val1", q[rnd].val1);
			sessionStorage.setItem("val2", q[rnd].val2);
			sessionStorage.setItem("odaimap", map.join(""));

			location = "name";
		} else{
			console.log(obj);
			alert("部屋の作成に失敗しました。");
		}
	});
}

function makeStr(chr, cnt){
	var str = "";
	for(let i = 0; i < cnt; i++)
		str += chr;
	return str;
}

function setMemberInput(){
	var names = document.getElementById("names");
	for(let i = 0; i < sessionStorage.getItem("playercount") - 0; i++){
		var count_div = document.createElement("div");
		count_div.setAttribute("class", "count_div");
		var input = document.createElement("input");
		input.setAttribute("type", "text");
		input.setAttribute("class", "count");
		input.name = "name" + i;
		input.id = "name" + i;
		input.setAttribute("maxlength", "50");
		var span = document.createElement("span");
		span.style.fontSize = "1.4em";
		span.style.userSelect = "none";
		span.innerHTML = "さん";
		count_div.appendChild(input);
		count_div.appendChild(span);
		names.appendChild(count_div);
	}
}

function name_next(){
	var error = false;
	for(let i = 0; i < sessionStorage.getItem("playercount") - 0; i++){
		if(document.getElementById("name" + i).value.length == 0){
			error = true;
		}
	}

	if(error){
		alert("全員の名前を入力してください。");
		return;
	}

	var members = new Object();
	for(let i = 0; i < sessionStorage.getItem("playercount") - 0; i++){
		members[i] = document.getElementById("name" + i).value;
	}
	sessionStorage.setItem("members", JSON.stringify(members));

	location = "game";
}

function odaiCheck(id){
	var members = JSON.parse(sessionStorage.getItem("members"));
	document.getElementById("msg").innerHTML = "";
	document.getElementById("msg2").innerHTML = members[id] + " さんですか？";
	document.getElementById("next").innerHTML = "はい";
	document.getElementById("nowId").value = id;
	mode = 0;
}

function nextMemberOdai(){
	if(mode == 0){
		var members = JSON.parse(sessionStorage.getItem("members"));
		var id = document.getElementById("nowId").value - 0;
		var odai = "";
		if(sessionStorage.getItem("odaimap").charAt(id) == 0){
			odai = sessionStorage.getItem("val1");
		} else{
			odai = sessionStorage.getItem("val2");
		}
		document.getElementById("msg").innerHTML = members[id] + " さんのお題は";
		document.getElementById("msg2").innerHTML = odai;
		document.getElementById("next").innerHTML = "OK";
		mode = 1;
	} else{
		var nextId = document.getElementById("nowId").value - 0 + 1;
		if(nextId == sessionStorage.getItem("playercount") - 0){
			startCheck();
		} else{
			odaiCheck(nextId);
		}
	}
}

function startCheck(){
	document.getElementById("msg2").innerHTML = "確認が終わりました。";
	document.getElementById("next").setAttribute("onclick", "location = 'offplay';");
	document.getElementById("next").innerHTML = "ゲームを開始する";
}

function timer(){
	var timerSec = sessionStorage.getItem("talktime") - 0;
	if(timerSec == 0) return;
	let startTime = new Date().getTime() / 1000;
	var time = document.getElementById("time");
	time.innerHTML = frontZero((timerSec / 60) | 0) + ":" + frontZero(timerSec % 60);
	var si = setInterval(() => {
		let currentTime = new Date().getTime() / 1000;
		let sec = timerSec - Math.floor(currentTime - startTime);
		time.innerHTML = frontZero((sec / 60) | 0) + ":" + frontZero(sec % 60);
		if(sec == 0){
			clearInterval(si);
			location = "finish";
		}
	}, 1000);
}

function frontZero(num){
	if(num < 10) return "0" + num;
	return num;
}

function setAnnounce(){
	var members = JSON.parse(sessionStorage.getItem("members"));
	for(let i = 0; i < sessionStorage.getItem("playercount") - 0; i++){
		var odai = "";
		if(sessionStorage.getItem("odaimap").charAt(i) == 0){
			odai = sessionStorage.getItem("val1");
		} else{
			odai = sessionStorage.getItem("val2");
		}
		var p = document.createElement("p");
		p.innerHTML = members[i] + " さんは " + odai + " です";
		document.getElementById("waitmember").appendChild(p);
	}
}

function rensen(){
	fetch('/r/allquestions', {
		method: "GET"
	}).then((res) => {
		return res.json();
	}).then((obj) => {
		if(obj.result_type == 0){
			var q = new Array();
			for(let i = 0; i < Object.keys(obj.question).length; i++){
				if(obj.question[i].category == sessionStorage.getItem("talkcategory"))
					q.push(obj.question[i]);
			}
			var rnd = Math.floor(Math.random() * Math.floor(q.length));
			var map = makeStr('0', (sessionStorage.getItem("playercount") - 0) - (sessionStorage.getItem("wordwolfcount") - 0)) + makeStr('1', (sessionStorage.getItem("wordwolfcount") - 0));

			sessionStorage.setItem("val1", q[rnd].val1);
			sessionStorage.setItem("val2", q[rnd].val2);
			sessionStorage.setItem("odaimap", map.split("").sort(() => Math.random() - 0.5).join(""));

			location = "game";
		} else{
			console.log(obj);
			alert("部屋の作成に失敗しました。");
		}
	});
}

function checkValueSpan(elm){
	var max = elm.getAttribute("max") - 0;
	var min = elm.getAttribute("min") - 0;
	if(elm.value - 0 < min) elm.value = min;
	else if(elm.value - 0 > max) elm.value = max;

	var talktime_min = document.querySelector("#talktime_min").value - 0;
	var talktime_sec = document.querySelector("#talktime_sec").value - 0;

	document.fm.talktime.value = "00:"
		+ (talktime_min < 10 ? "0" : "") + talktime_min + ":"
		+ (talktime_sec < 10 ? "0" : "") + talktime_sec;
}