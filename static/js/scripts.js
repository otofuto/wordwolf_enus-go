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
		document.body.style.padding = "0 34%";
	}
	else{ //縦画面の場合
		document.body.style.padding = "0";
	}
}

function roommake_next(){
	var error = false;
	if(document.fm.username.value.length == 0){
		document.getElementById("error_username").style.display = "inline-block";
		error = true;
	}
	else document.getElementById("error_username").style.display = "none";

	if(document.fm.roomname.value.length == 0){
		document.getElementById("error_roomname1").style.display = "inline-block";
		error = true;
	}
	else document.getElementById("error_roomname1").style.display = "none";

	if(document.fm.roomname.value.length > 10){
		document.getElementById("error_roomname2").style.display = "inline-block";
		error = true;
	}
	else document.getElementById("error_roomname2").style.display = "none";

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

	if(document.fm.passcode.value.length == 0){
		document.getElementById("error_pass1").style.display = "inline-block";
		error = true;
	}
	else document.getElementById("error_pass1").style.display = "none";

	if(document.fm.passcode.value.length > 0 && document.fm.passcode.value.length != 4){
		document.getElementById("error_pass2").style.display = "inline-block";
		error = true;
	}
	else document.getElementById("error_pass2").style.display = "none";

	if(error){
		alert("入力内容に誤りがあります。");
		return;
	}

	localStorage.setItem('wf_myname', document.fm.username.value);
	
	document.fm.submit();
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

function goSetting(){
	fetch('/r/setting', {
		method: "get"
	}).then(res => {
		location = 'setting';
	}).catch(ex => {
		alert("オフラインではこの機能はご利用になれません。");
	});
}