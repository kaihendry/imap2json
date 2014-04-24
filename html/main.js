var mailjson = {};

function threadview(id) {
	// console.log(id);
	console.time("Finding thread");
	var arg = id.split('-'); // hash of first message - UID
	id = arg[0];
	//console.log(id);

	// I have no idea why I wrote this
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

	// console.log(mailjson[i].Id);
	if (typeof mailjson[i] === 'undefined') {
		$('#title').html("Thread not found.");
		return;
	}

	// console.log("Found ", id, " at ", i);
	$('#title').html(mailjson[i].Msgs[0].Header.Subject);

	console.time('Rendering thread');

	$.getJSON("c/" + mailjson[i].Id + ".json", function( data ) {
		// console.log(data);


	$.each(data.Msgs, function(index, value) {
		// console.log(value);
		msg = "<div class=mail>";
		// TODO make permalinks to individual messages work
		msg += '<a title="UID" id=' + id + '-' + value.UID + ' href=#' + id + '-' + value.UID + ' class="uid">' + value.UID + '</a>';

		msg += '<dl>';
		msg += '<dt>From:</dt><dd>';
		for (f in value.Header.From) {
			msg += '<strong>' + value.Header.From[f].Name + '</strong>&lt;' + value.Header.From[f].Address + '&gt;';
		}
		msg += '</dd>';
		msg += '<dt>To:</dt><dd>';
		for (f in value.Header.To) {
			msg += '<strong>' + value.Header.To[f].Name + '</strong>&lt;' + value.Header.To[f].Address + '&gt;';
		}
		msg += '</dd>';
		msg += '<dt>Time:</dt><dd><time class="time">' + value.Date + '</time></dd>';
		msg += '</dl>';
		msg += ' <span><a title="A url to download the original RFC8222 message from" href=raw/' + value.UID + '.txt>rawUrl</a></span>';
		msg += "<hr><pre>";
		msg += $("<pre/>").text(value.Body).html();
		msg += "</pre></div>";
		$("#conversation").append(msg);
		console.timeEnd('Rendering thread');
		// console.log(index + ": " + value);
	});



	});

}

function main() {
	id = window.location.hash.substr(1)
	if (id) {
		threadview(id);
	} else {
		console.time('Rendering index');
		$("#title").html('imap2json of ' + mailjson.length + " conversations");
		$.each(mailjson, function(index, value) {
			// console.log(value.Id)
			try {
				var c = "<a href=#" ;
				c += value.Id + ">";
				c += "<span class='count'>" + value.Count + "</span>";
				for (var f in value.Msgs[0].Header.From) {
				c += "<span class=from>" + value.Msgs[0].Header.From[f].Name + "</span>";
				}
				//for (var f in value.Msgs[0].Header.To) {
				//	if (value.Msgs[0].Header.To[f].Name) {
				//c += "<span class=to>" + value.Msgs[0].Header.To[f].Name + "</span>";
				//	} else {
				//c += "<span class=to>&lt;" + value.Msgs[0].Header.To[f].Address + "&gt;</span>";
				//	}
				// }
				c += "<span class=subject>" +value.Msgs[0].Header.Subject + "</span>";
				c += "<span><time>" + value.Msgs[0].Date + "</time></span>";
				c += '</a>'

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
		mailjson = data;
		main();
	});

	$(window).bind('hashchange', function() {
		$("#title").html('');
		$("#conversation").html('');
		main();
	});

});

