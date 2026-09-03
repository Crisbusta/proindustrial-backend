-- El recorte de descripciones no es reversible: el texto sobrante ya no
-- está en la base. Marcar los CHECK como "no validados" otra vez tampoco
-- aportaría nada, porque NOT VALID no exime a los UPDATE posteriores
-- (ese fue justamente el error que 012 corrige). Para soltar los topes,
-- el camino es el down de 011, que los elimina.
DO $$
BEGIN
  RAISE NOTICE '012 down: sin efecto; el recorte de descripciones es irreversible.';
END $$;
