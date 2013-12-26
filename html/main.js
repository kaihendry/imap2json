var mailjson = {};

function threadview(id) {
	initial = mailjson.THREAD[id][0];
	$('#title').html(mailjson[initial].Header.Subject);
	$.each(mailjson.THREAD[id], function(index, value) {
		msg = "<div class=mail>";
		msg += '<span class="from">';
		msg += '<span class="name"><span>' + mailjson[value].Header.From + '</span>'
		msg += '</span>';
		msg += '<span class="to">to ';
		msg += '<span class="name"><span>' + mailjson[value].Header.To + '</span>'
		msg += '</span><br>';
		msg += '<time class="time">' + mailjson[value].Header.Date + '</time>';
		msg += "<hr><pre>" + mailjson[value].Body + "</pre></div>";
		$("#conversation").append(msg);
		console.log(index + ": " + value);
	});
}

function main() {
	id = window.location.hash.substr(1)
	if (id) {
		threadview(id);
	} else {
		$("#title").html('mail2json Index');
		$.each(mailjson.THREAD, function(index, value) {
			try {
			$("#conversation").append("<li><strong><a href=#" + index + ">" + index + "</strong> " + value + " Subject: " + mailjson[value[0]].Header.Subject + "</a></li>");
			} catch (e) {
				console.log(value, e);
			}
		});
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

