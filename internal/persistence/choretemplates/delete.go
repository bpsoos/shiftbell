package choretemplates

import "fmt"

func (p *Persister) Delete(id int) error {
	_, err := p.db.Exec(`
		delete from chore_templates where id = ?
	`,
		id,
	)
	if err != nil {
		return fmt.Errorf("delete exec: %v", err)
	}
	return nil
}
