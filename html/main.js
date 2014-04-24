var mailjson = {};

function threadview(id) {
	// console.log(id);
	console.time("Finding thread");
	var arg = id.split('-'); // hash of first message - UID
	id = arg[0];
	//console.log(id);
	if (!parseInt(arg[1]) > 0) {
		window.scrollTo(0, 0);
		console.log("Scrolling to top");
	}
	for (i = 0; i < mailjson.length; i++) {
		// console.log(id);
		if (mailjson[i].Id.indexOf(id) == 0) {
			break;
		}
	}
	console.timeEnd('Finding thread');

	// console.log(i);
	if (typeof mailjson[i] === 'undefined') {
		$('#title').html("Thread not found.");
		return;
	}

	// console.log("Found ", id, " at ", i);
	$('#title').html(mailjson[i].Msgs[0].Header.Subject);

	console.time('Rendering thread');

	$.each(mailjson[i].Msgs, function(index, value) {
		msg = "<div class=mail>";
		msg += '<a title="UID" id=' + id + '-' + value.UID + ' href=#' + id + '-' + value.UID + ' class="uid">' + value.UID + '</a>';
		msg += '<span class="from">';
		msg += '<span class="name"><span>' + value.Header.From + '</span>'
		msg += '</span>';
		msg += '<span class="to">to ';
		msg += '<span class="name"><span>' + value.Header.To + '</span>'
		msg += '</span><br>';
		msg += '<time class="time">' + value.Header.Date + '</time>';
		msg += ' <span><a title="A url to download the original RFC8222 message from" href=raw/' + value.UID + '.txt>rawUrl</a></span>';
		msg += ' <span><a href=https://github.com/kaihendry/imap2json/issues>Notice an issue?</a></span>';
		msg += "<hr><pre>";
		msg += $("<pre/>").text(value.Body).html();
		msg += "</pre></div>";
		$("#conversation").append(msg);
		console.timeEnd('Rendering thread');
		// console.log(index + ": " + value);
	});
}

function main() {
	id = window.location.hash.substr(1)
	if (id) {
		threadview(id);
	} else {
		console.time('Rendering index');
		$("#title").html('mail2json Index');
		$.each(mailjson, function(index, value) {
			// console.log(value.Id)
			try {
				var c = "<li><a href=#" ;
				c += value.Id + ">";
				if (value.Msgs.length == 1) {
					c += "<span class='thread1'>1</span>";
				} else {
					c += "<span class='threadconv'>" + value.Msgs.length + "</span>";
				}
				c += "</span> <time>" + value.Msgs[0].Header.Date + "</time>&nbsp;<strong>" +value.Msgs[0].Header.Subject + "</strong></a></li>";
				$("#conversation").append(c);
			} catch(e) {
				console.log(value, e);
			}
		});
		console.timeEnd('Rendering index');
	}
}

$(function() {
	$.getJSON("mail.json").done(function(data) {
		mailjson = data.reverse();
		main();
	});

	$(window).bind('hashchange', function() {
		$("#title").html('');
		$("#conversation").html('');
		main();
	});

});

