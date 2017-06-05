# phoru
Given a (pseudo-)phonetic russian input, phoru outputs (cyrillic) russian.

	% phoru -h
	
	example:
		 % echo privet mir | phoru
		 % привет мир
	
	translation tables:
		 a : а
		 b : б
		 d : д
		 e : е
		 f : ф
		 g : г
		 i : и
		 j : ж
		 k : к
		 l : л
		 m : м
		 n : н
		 o : о
		 p : п
		 r : р
		 s : с
		 t : т
		 u : у
		 v : в
		 z : з
		 è : э
		 î : ы
		 ï : й
		 `e : э
		 b- : ь
		 ch : ч
		 i- : ы
		 i_ : й
		 kh : х
		 sh : ш
		 ya : я
		 yo : ё
		 shh : щ
		 ts- : ц
		 you : ю
	
	in server mode:
		 phoru -http=:6060
	  -h	show this help
	  -http string
	    	Run in http server mode, on the given address.
	  -v	verbose
	
	